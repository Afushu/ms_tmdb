package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"ms_tmdb/config"
	"ms_tmdb/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	defaultRetentionDays    = 7
	defaultBodyLimitBytes   = 64 * 1024
	defaultWriteTimeout     = 3 * time.Second
	defaultCleanupBatchSize = 500
	proxyAccessLogTableName = "proxy_access_logs"
	tmdbRequestLogTableName = "tmdb_request_logs"
)

// BodySnapshot 保存被截断后的正文与原始大小。
type BodySnapshot struct {
	Text      string
	Bytes     int64
	Truncated bool
}

// ProxyAccessEntry 是一次外部代理访问的落库载荷。
type ProxyAccessEntry struct {
	RequestID string
	Method    string

	Path       string
	Query      string
	RequestURI string
	ClientIP   string
	UserAgent  string

	StatusCode   int
	DurationMs   int64
	ErrorMessage string

	RequestBody  BodySnapshot
	ResponseBody BodySnapshot
}

// TmdbRequestEntry 是一次真实 TMDB 上游请求的落库载荷。
type TmdbRequestEntry struct {
	RequestID string
	Method    string

	Path string
	URL  string

	StatusCode   int
	DurationMs   int64
	ErrorMessage string

	RequestBody  BodySnapshot
	ResponseBody BodySnapshot
}

// RequestLogService 负责请求日志的截断、写入和保留期清理。
type RequestLogService struct {
	db             *gorm.DB
	retentionDays  int
	bodyLimitBytes int
}

// NewRequestLogService 创建日志服务，并归一化缺省配置。
func NewRequestLogService(db *gorm.DB, c config.TmdbLogConf) *RequestLogService {
	retentionDays := c.RetentionDays
	if retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}

	bodyLimitBytes := c.BodyLimitBytes
	if bodyLimitBytes <= 0 {
		bodyLimitBytes = defaultBodyLimitBytes
	}

	return &RequestLogService{
		db:             db,
		retentionDays:  retentionDays,
		bodyLimitBytes: bodyLimitBytes,
	}
}

// BodyLimitBytes 返回单个正文允许保存的最大字节数。
func (s *RequestLogService) BodyLimitBytes() int {
	if s == nil || s.bodyLimitBytes <= 0 {
		return defaultBodyLimitBytes
	}
	return s.bodyLimitBytes
}

// CaptureBody 按配置截断正文，数据库中只保存可读文本。
func (s *RequestLogService) CaptureBody(raw []byte) BodySnapshot {
	limit := defaultBodyLimitBytes
	if s != nil && s.bodyLimitBytes > 0 {
		limit = s.bodyLimitBytes
	}
	return CaptureBody(raw, limit)
}

// WriteProxyAccessAsync 异步写入外部访问日志，避免日志库慢时拖慢代理响应。
func (s *RequestLogService) WriteProxyAccessAsync(ctx context.Context, entry ProxyAccessEntry) {
	if s == nil || s.db == nil {
		return
	}

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				logx.Errorf("写入代理访问日志异常: %v", recovered)
			}
		}()

		writeCtx, cancel := context.WithTimeout(withoutCancel(ctx), defaultWriteTimeout)
		defer cancel()

		if err := s.WriteProxyAccess(writeCtx, entry); err != nil {
			logx.Errorf("写入代理访问日志失败: %v", err)
		}
	}()
}

// CaptureBody 按指定上限截断正文。
func CaptureBody(raw []byte, limit int) BodySnapshot {
	if limit <= 0 {
		limit = defaultBodyLimitBytes
	}

	size := int64(len(raw))
	textBytes := raw
	truncated := len(raw) > limit
	if truncated {
		textBytes = raw[:limit]
	}

	return BodySnapshot{
		Text:      strings.ToValidUTF8(string(textBytes), "?"),
		Bytes:     size,
		Truncated: truncated,
	}
}

// WriteProxyAccess 写入外部访问日志。
func (s *RequestLogService) WriteProxyAccess(ctx context.Context, entry ProxyAccessEntry) error {
	if s == nil || s.db == nil {
		return nil
	}

	mediaType, tmdbID := parseProxyMediaTarget(entry.Path)
	return s.db.WithContext(contextOrBackground(ctx)).Create(&model.ProxyAccessLog{
		RequestID: entry.RequestID,
		Method:    entry.Method,

		MediaType: mediaType,
		TmdbID:    tmdbID,

		Path:       entry.Path,
		Query:      entry.Query,
		RequestURI: entry.RequestURI,
		ClientIP:   entry.ClientIP,
		UserAgent:  entry.UserAgent,

		StatusCode:   entry.StatusCode,
		DurationMs:   entry.DurationMs,
		ErrorMessage: entry.ErrorMessage,

		RequestBody:           entry.RequestBody.Text,
		RequestBodyBytes:      entry.RequestBody.Bytes,
		RequestBodyTruncated:  entry.RequestBody.Truncated,
		ResponseBody:          entry.ResponseBody.Text,
		ResponseBodyBytes:     entry.ResponseBody.Bytes,
		ResponseBodyTruncated: entry.ResponseBody.Truncated,
	}).Error
}

func parseProxyMediaTarget(path string) (string, int) {
	tmdbPath := strings.TrimSpace(path)
	if tmdbPath == "" {
		return "", 0
	}

	for _, prefix := range []string{"/api/v3", "/v3", "/3"} {
		if strings.HasPrefix(tmdbPath, prefix) {
			tmdbPath = strings.TrimPrefix(tmdbPath, prefix)
			break
		}
	}

	parts := strings.Split(strings.Trim(tmdbPath, "/"), "/")
	if len(parts) < 2 {
		return "", 0
	}

	mediaType := parts[0]
	if mediaType != "movie" && mediaType != "tv" {
		return "", 0
	}

	tmdbID, err := strconv.Atoi(parts[1])
	if err != nil || tmdbID == 0 {
		return "", 0
	}
	return mediaType, tmdbID
}

// WriteTmdbRequest 写入真实 TMDB 上游请求日志。
func (s *RequestLogService) WriteTmdbRequest(ctx context.Context, entry TmdbRequestEntry) error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.WithContext(contextOrBackground(ctx)).Create(&model.TmdbRequestLog{
		RequestID: entry.RequestID,
		Method:    entry.Method,

		Path: entry.Path,
		URL:  entry.URL,

		StatusCode:   entry.StatusCode,
		DurationMs:   entry.DurationMs,
		ErrorMessage: entry.ErrorMessage,

		RequestBody:           entry.RequestBody.Text,
		RequestBodyBytes:      entry.RequestBody.Bytes,
		RequestBodyTruncated:  entry.RequestBody.Truncated,
		ResponseBody:          entry.ResponseBody.Text,
		ResponseBodyBytes:     entry.ResponseBody.Bytes,
		ResponseBodyTruncated: entry.ResponseBody.Truncated,
	}).Error
}

// CleanupExpired 物理删除超过保留期的日志。
func (s *RequestLogService) CleanupExpired(ctx context.Context) error {
	if s == nil || s.db == nil {
		return nil
	}

	retentionDays := s.retentionDays
	if retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	db := s.db.WithContext(withoutCancel(ctx))
	if err := cleanupExpiredRequestLogTable(db, proxyAccessLogTableName, cutoff); err != nil {
		return err
	}
	return cleanupExpiredRequestLogTable(db, tmdbRequestLogTableName, cutoff)
}

func cleanupExpiredRequestLogTable(db *gorm.DB, tableName string, cutoff time.Time) error {
	for {
		result := db.Exec(`
WITH expired AS (
  SELECT id
  FROM `+tableName+`
  WHERE created_at < ?
  ORDER BY created_at ASC, id ASC
  LIMIT ?
)
DELETE FROM `+tableName+`
WHERE id IN (SELECT id FROM expired)
`, cutoff, defaultCleanupBatchSize)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected < int64(defaultCleanupBatchSize) {
			return nil
		}
	}
}

// StartRetentionCleaner 启动每日一次的日志保留期清理。
func (s *RequestLogService) StartRetentionCleaner(ctx context.Context) func() {
	cleanerCtx, cancel := context.WithCancel(withoutCancel(ctx))

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-cleanerCtx.Done():
				return
			case <-ticker.C:
				if err := s.CleanupExpired(cleanerCtx); err != nil {
					logx.Errorf("清理请求日志失败: %v", err)
				}
			}
		}
	}()

	return cancel
}

func withoutCancel(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
