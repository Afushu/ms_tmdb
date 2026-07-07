package admin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	minRequestTimeoutMillis = int64(1000)
	maxRequestTimeoutMillis = int64(300000)
)

type UpdateProxySettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProxySettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProxySettingsLogic {
	return &UpdateProxySettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProxySettingsLogic) UpdateProxySettings(req *types.AdminProxyReq) (*types.AdminProxyResp, error) {
	oldProxyURL := l.svcCtx.TmdbClient.GetProxy()
	oldLocalWriteEnabled := l.svcCtx.ProxyService.LocalWriteEnabled()
	oldTimeout := l.svcCtx.Config.Timeout

	proxyURL := oldProxyURL
	if req.ProxyURL != nil {
		proxyURL = strings.TrimSpace(*req.ProxyURL)
	}

	localWriteEnabled := oldLocalWriteEnabled
	if req.LocalWriteEnabled != nil {
		localWriteEnabled = *req.LocalWriteEnabled
	}
	timeout := oldTimeout
	if req.Timeout != nil {
		if err := validateRequestTimeout(*req.Timeout); err != nil {
			return nil, err
		}
		timeout = *req.Timeout
	}

	if err := l.svcCtx.TmdbClient.SetProxy(proxyURL); err != nil {
		return nil, err
	}
	l.svcCtx.TmdbClient.SetTimeoutMillis(timeout)
	l.svcCtx.ProxyService.SetLocalWriteEnabled(localWriteEnabled)
	l.svcCtx.Config.Tmdb.LocalWriteEnabled = localWriteEnabled
	l.svcCtx.Config.Timeout = timeout

	configFile := strings.TrimSpace(l.svcCtx.Config.ConfigFile)
	if configFile == "" {
		configFile = "etc/tmdb.yaml"
	}
	if err := writeProxySettingsToConfigFile(configFile, proxyURL, localWriteEnabled, req.Timeout); err != nil {
		// 配置写入失败时回滚当前进程设置，避免“显示成功但重启丢失”。
		_ = l.svcCtx.TmdbClient.SetProxy(oldProxyURL)
		l.svcCtx.TmdbClient.SetTimeoutMillis(oldTimeout)
		l.svcCtx.ProxyService.SetLocalWriteEnabled(oldLocalWriteEnabled)
		l.svcCtx.Config.Tmdb.LocalWriteEnabled = oldLocalWriteEnabled
		l.svcCtx.Config.Timeout = oldTimeout
		return nil, err
	}

	return &types.AdminProxyResp{
		ProxyURL:               proxyURL,
		Enabled:                proxyURL != "",
		LocalWriteEnabled:      localWriteEnabled,
		Timeout:                timeout,
		TimeoutRestartRequired: timeoutRestartRequired(l.svcCtx),
	}, nil
}

func validateRequestTimeout(timeout int64) error {
	if timeout < minRequestTimeoutMillis || timeout > maxRequestTimeoutMillis {
		return fmt.Errorf("请求超时时间必须在 %d 到 %d 毫秒之间", minRequestTimeoutMillis, maxRequestTimeoutMillis)
	}
	return nil
}

func timeoutRestartRequired(svcCtx *svc.ServiceContext) bool {
	if svcCtx == nil {
		return false
	}
	return svcCtx.Config.Timeout != svcCtx.StartupTimeout
}

func writeProxySettingsToConfigFile(configPath string, proxyURL string, localWriteEnabled bool, timeout *int64) error {
	configFile, err := readConfigFileLines(configPath)
	if err != nil {
		return err
	}

	configFile.lines, err = applyTmdbConfigValues(configFile.lines, map[string]string{
		"ProxyURL":          yamlDoubleQuoted(proxyURL),
		"LocalWriteEnabled": fmt.Sprintf("%t", localWriteEnabled),
	}, []string{"ProxyURL", "LocalWriteEnabled"})
	if err != nil {
		return err
	}
	if timeout != nil {
		configFile.lines = applyTopLevelConfigValues(configFile.lines, map[string]string{
			"Timeout": fmt.Sprintf("%d", *timeout),
		}, []string{"Timeout"})
	}
	return configFile.write()
}

type configFileLines struct {
	path    string
	lines   []string
	lineSep string
}

func readConfigFileLines(configPath string) (*configFileLines, error) {
	configFilePath := filepath.Clean(strings.TrimSpace(configPath))
	if configFilePath == "" {
		return nil, errors.New("配置文件路径为空")
	}

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	raw := string(content)
	lineSep := "\n"
	if strings.Contains(raw, "\r\n") {
		lineSep = "\r\n"
		raw = strings.ReplaceAll(raw, "\r\n", "\n")
	}

	return &configFileLines{
		path:    configFilePath,
		lines:   strings.Split(raw, "\n"),
		lineSep: lineSep,
	}, nil
}

func (c *configFileLines) write() error {
	output := strings.Join(c.lines, "\n")
	if c.lineSep == "\r\n" {
		output = strings.ReplaceAll(output, "\n", "\r\n")
	}
	if err := os.WriteFile(c.path, []byte(output), 0o644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

func applyTmdbConfigValues(lines []string, values map[string]string, order []string) ([]string, error) {
	tmdbFound := false
	inTmdb := false
	tmdbIndent := 0
	tmdbStartIndex := -1
	tmdbEndIndex := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		indent := leadingIndentLen(line)

		if !inTmdb {
			if strings.HasPrefix(trimmed, "Tmdb:") {
				tmdbFound = true
				inTmdb = true
				tmdbIndent = indent
				tmdbStartIndex = i
			}
			continue
		}

		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && indent <= tmdbIndent {
			tmdbEndIndex = i
			break
		}
	}

	if !tmdbFound {
		return nil, errors.New("配置文件缺少 Tmdb 段")
	}
	if tmdbEndIndex < 0 {
		tmdbEndIndex = len(lines)
	}

	lineIndexes := make(map[string]int, len(values))
	autoSyncIndex := -1
	for i := tmdbStartIndex + 1; i < tmdbEndIndex; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if autoSyncIndex < 0 && strings.HasPrefix(trimmed, "AutoSync:") {
			autoSyncIndex = i
		}
		for key := range values {
			if strings.HasPrefix(trimmed, key+":") {
				lineIndexes[key] = i
			}
		}
	}

	for _, key := range order {
		value, ok := values[key]
		if !ok {
			continue
		}
		line := strings.Repeat(" ", tmdbIndent+2) + key + ": " + value
		if index, exists := lineIndexes[key]; exists {
			lines[index] = line
			continue
		}

		insertIndex := tmdbEndIndex
		if key == "LocalWriteEnabled" {
			if proxyIndex, exists := lineIndexes["ProxyURL"]; exists {
				insertIndex = proxyIndex + 1
			} else if autoSyncIndex >= 0 {
				insertIndex = autoSyncIndex
			}
		} else if autoSyncIndex >= 0 {
			insertIndex = autoSyncIndex
		}

		lines = append(lines[:insertIndex], append([]string{line}, lines[insertIndex:]...)...)
		tmdbEndIndex++
		for existingKey, index := range lineIndexes {
			if index >= insertIndex {
				lineIndexes[existingKey] = index + 1
			}
		}
		if autoSyncIndex >= insertIndex {
			autoSyncIndex++
		}
		lineIndexes[key] = insertIndex
	}

	return lines, nil
}

func applyTopLevelConfigValues(lines []string, values map[string]string, order []string) []string {
	lineIndexes := make(map[string]int, len(values))
	insertIndex := len(lines)
	portIndex := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || leadingIndentLen(line) != 0 {
			continue
		}
		if strings.HasPrefix(trimmed, "Port:") {
			portIndex = i
		}
		if strings.HasPrefix(trimmed, "Postgres:") || strings.HasPrefix(trimmed, "Tmdb:") {
			insertIndex = i
			break
		}
		for key := range values {
			if strings.HasPrefix(trimmed, key+":") {
				lineIndexes[key] = i
			}
		}
	}
	if portIndex >= 0 {
		insertIndex = portIndex + 1
	}

	for _, key := range order {
		value, ok := values[key]
		if !ok {
			continue
		}
		line := key + ": " + value
		if index, exists := lineIndexes[key]; exists {
			lines[index] = line
			continue
		}

		lines = append(lines[:insertIndex], append([]string{line}, lines[insertIndex:]...)...)
		for existingKey, index := range lineIndexes {
			if index >= insertIndex {
				lineIndexes[existingKey] = index + 1
			}
		}
		lineIndexes[key] = insertIndex
		insertIndex++
	}

	return lines
}

func leadingIndentLen(line string) int {
	length := 0
	for _, ch := range line {
		if ch != ' ' && ch != '\t' {
			break
		}
		length++
	}
	return length
}

func yamlDoubleQuoted(text string) string {
	escaped := strings.ReplaceAll(text, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
