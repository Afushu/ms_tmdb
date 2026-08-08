// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package admin

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	minLogRetentionDays = 1
	maxLogRetentionDays = 365
)

type UpdateLogSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLogSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogSettingsLogic {
	return &UpdateLogSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLogSettingsLogic) UpdateLogSettings(req *types.AdminLogSettingsReq) (*types.AdminLogSettingsResp, error) {
	if req == nil {
		return nil, errors.New("请求参数不能为空")
	}

	oldRetentionDays := currentLogRetentionDays(l.svcCtx)
	nextRetentionDays := oldRetentionDays
	if req.RetentionDays != nil {
		if err := validateLogRetentionDays(*req.RetentionDays); err != nil {
			return nil, err
		}
		nextRetentionDays = *req.RetentionDays
	}

	if l.svcCtx.LogService != nil {
		l.svcCtx.LogService.SetRetentionDays(nextRetentionDays)
	}
	l.svcCtx.Config.Tmdb.Log.RetentionDays = nextRetentionDays

	configFile := strings.TrimSpace(l.svcCtx.Config.ConfigFile)
	if configFile == "" {
		configFile = "etc/tmdb.yaml"
	}
	if err := writeLogSettingsToConfigFile(configFile, nextRetentionDays, l.svcCtx.Config.Tmdb.Log.BodyLimitBytes); err != nil {
		// 配置写入失败时回滚当前进程设置，避免“显示成功但重启丢失”。
		if l.svcCtx.LogService != nil {
			l.svcCtx.LogService.SetRetentionDays(oldRetentionDays)
		}
		l.svcCtx.Config.Tmdb.Log.RetentionDays = oldRetentionDays
		return nil, err
	}

	return &types.AdminLogSettingsResp{
		RetentionDays: currentLogRetentionDays(l.svcCtx),
	}, nil
}

func currentLogRetentionDays(svcCtx *svc.ServiceContext) int {
	if svcCtx == nil {
		return 7
	}
	if svcCtx.LogService != nil {
		return svcCtx.LogService.RetentionDays()
	}
	if svcCtx.Config.Tmdb.Log.RetentionDays > 0 {
		return svcCtx.Config.Tmdb.Log.RetentionDays
	}
	return 7
}

func validateLogRetentionDays(days int) error {
	if days < minLogRetentionDays || days > maxLogRetentionDays {
		return fmt.Errorf("日志保留天数必须在 %d 到 %d 天之间", minLogRetentionDays, maxLogRetentionDays)
	}
	return nil
}

// writeLogSettingsToConfigFile 将日志保留配置写入 yaml，并尽量保留既有 BodyLimitBytes。
func writeLogSettingsToConfigFile(configPath string, retentionDays int, bodyLimitBytes int) error {
	configFile, err := readConfigFileLines(configPath)
	if err != nil {
		return err
	}

	configFile.lines, err = applyLogSettingsConfigValues(configFile.lines, retentionDays, bodyLimitBytes)
	if err != nil {
		return err
	}
	return configFile.write()
}

func applyLogSettingsConfigValues(lines []string, retentionDays int, bodyLimitBytes int) ([]string, error) {
	tmdbFound := false
	tmdbStart := -1
	tmdbIndent := 0
	tmdbEnd := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		indent := leadingIndentLen(line)

		if !tmdbFound {
			if strings.HasPrefix(trimmed, "Tmdb:") {
				tmdbFound = true
				tmdbStart = i
				tmdbIndent = indent
			}
			continue
		}

		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && indent <= tmdbIndent {
			tmdbEnd = i
			break
		}
	}
	if !tmdbFound {
		return nil, errors.New("配置文件缺少 Tmdb 段")
	}

	parentIndent := tmdbIndent + 2
	logStart := -1
	logEnd := tmdbEnd
	existingBodyLimit := bodyLimitBytes
	for i := tmdbStart + 1; i < tmdbEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		indent := leadingIndentLen(lines[i])

		if strings.HasPrefix(trimmed, "Log:") && indent == parentIndent {
			logStart = i
			logEnd = tmdbEnd
			for j := i + 1; j < tmdbEnd; j++ {
				childTrimmed := strings.TrimSpace(lines[j])
				childIndent := leadingIndentLen(lines[j])
				if childTrimmed != "" && !strings.HasPrefix(childTrimmed, "#") && childIndent <= parentIndent {
					logEnd = j
					break
				}
				if strings.HasPrefix(childTrimmed, "BodyLimitBytes:") {
					if value, ok := parseYAMLIntValue(childTrimmed); ok && value > 0 {
						existingBodyLimit = value
					}
				}
			}
			break
		}
	}

	indent := strings.Repeat(" ", parentIndent)
	childIndent := strings.Repeat(" ", parentIndent+2)
	block := []string{
		indent + "Log:",
		childIndent + "RetentionDays: " + strconv.Itoa(retentionDays),
	}
	if existingBodyLimit > 0 {
		block = append(block, childIndent+"BodyLimitBytes: "+strconv.Itoa(existingBodyLimit))
	}

	if logStart >= 0 {
		lines = append(lines[:logStart], append(block, lines[logEnd:]...)...)
	} else {
		lines = append(lines[:tmdbEnd], append(block, lines[tmdbEnd:]...)...)
	}
	return lines, nil
}

func parseYAMLIntValue(line string) (int, bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return 0, false
	}
	raw := strings.TrimSpace(parts[1])
	raw = strings.Trim(raw, `"'`)
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return value, true
}
