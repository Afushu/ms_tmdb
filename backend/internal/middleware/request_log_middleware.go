package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ms_tmdb/internal/service"

	"github.com/zeromicro/go-zero/core/logx"
)

type requestIDContextKey struct{}

// RequestLogMiddleware 统一记录 TMDB 代理入口的外部访问日志。
type RequestLogMiddleware struct {
	LogService *service.RequestLogService
}

func NewRequestLogMiddleware(logService *service.RequestLogService) *RequestLogMiddleware {
	return &RequestLogMiddleware{
		LogService: logService,
	}
}

func (m *RequestLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m == nil || m.LogService == nil {
			next(w, r)
			return
		}
		if _, ok := resolveTmdbPath(r.URL.Path); !ok {
			next(w, r)
			return
		}

		requestID := requestIDFromHeader(r)
		if requestID == "" {
			requestID = newProxyRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)

		startedAt := time.Now()
		requestBody := m.captureRequestBody(r)
		recorder := newLogResponseRecorder(w, m.bodyLimitBytes())

		next(recorder, r.WithContext(context.WithValue(r.Context(), requestIDContextKey{}, requestID)))

		m.writeProxyAccessLog(
			r,
			requestID,
			requestBody,
			recorder.BodySnapshot(),
			recorder.StatusCode(),
			time.Since(startedAt),
		)
	}
}

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(requestIDContextKey{}).(string)
	return strings.TrimSpace(value)
}

func requestIDFromHeader(r *http.Request) string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.Header.Get("X-Request-Id"))
}

func (m *RequestLogMiddleware) bodyLimitBytes() int {
	if m == nil || m.LogService == nil {
		return 64 * 1024
	}
	return m.LogService.BodyLimitBytes()
}

func (m *RequestLogMiddleware) captureRequestBody(r *http.Request) service.BodySnapshot {
	if r == nil || r.Body == nil {
		return service.CaptureBody(nil, m.bodyLimitBytes())
	}

	limit := m.bodyLimitBytes()
	body, err := io.ReadAll(io.LimitReader(r.Body, int64(limit)+1))
	if err != nil {
		logx.Errorf("读取代理请求正文失败: %v", err)
	}
	r.Body = readCloser{
		Reader: io.MultiReader(bytes.NewReader(body), r.Body),
		Closer: r.Body,
	}
	return requestBodySnapshot(body, r.ContentLength, limit)
}

func (m *RequestLogMiddleware) writeProxyAccessLog(
	r *http.Request,
	requestID string,
	requestBody service.BodySnapshot,
	responseBody service.BodySnapshot,
	statusCode int,
	duration time.Duration,
) {
	if m == nil || m.LogService == nil {
		return
	}

	query := sanitizeRawQuery(r.URL.RawQuery)
	requestURI := r.URL.Path
	if query != "" {
		requestURI += "?" + query
	}

	m.LogService.WriteProxyAccessAsync(r.Context(), service.ProxyAccessEntry{
		RequestID: requestID,
		Method:    r.Method,

		Path:       r.URL.Path,
		Query:      query,
		RequestURI: requestURI,
		ClientIP:   clientIP(r),
		UserAgent:  r.UserAgent(),

		StatusCode:   statusCode,
		DurationMs:   duration.Milliseconds(),
		ErrorMessage: responseErrorMessage(statusCode, responseBody.Text),

		RequestBody:  requestBody,
		ResponseBody: responseBody,
	})
}

func requestBodySnapshot(body []byte, contentLength int64, limit int) service.BodySnapshot {
	if limit <= 0 {
		limit = 64 * 1024
	}

	raw := body
	truncated := len(body) > limit
	if truncated {
		raw = body[:limit]
	}

	size := int64(len(body))
	if contentLength >= 0 {
		size = contentLength
		truncated = truncated || contentLength > int64(limit)
	}

	return service.BodySnapshot{
		Text:      strings.ToValidUTF8(string(raw), "?"),
		Bytes:     size,
		Truncated: truncated,
	}
}

func responseErrorMessage(statusCode int, body string) string {
	if statusCode >= 200 && statusCode < 400 {
		return ""
	}

	var payload struct {
		StatusMessage string `json:"status_message"`
		Message       string `json:"message"`
		Msg           string `json:"msg"`
		Error         string `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err == nil {
		for _, value := range []string{payload.StatusMessage, payload.Message, payload.Msg, payload.Error} {
			if text := strings.TrimSpace(value); text != "" {
				return text
			}
		}
	}

	if statusCode > 0 {
		if text := http.StatusText(statusCode); text != "" {
			return text
		}
	}
	return "代理请求失败"
}

func sanitizeRawQuery(raw string) string {
	if raw == "" {
		return ""
	}

	values, err := url.ParseQuery(raw)
	if err != nil {
		return maskAPIKeyInRawQuery(raw)
	}
	for key := range values {
		if strings.EqualFold(key, "api_key") {
			values.Set(key, "***")
		}
	}
	return values.Encode()
}

func maskAPIKeyInRawQuery(raw string) string {
	parts := strings.Split(raw, "&")
	for i, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		if strings.EqualFold(key, "api_key") {
			parts[i] = key + "=***"
			continue
		}
		parts[i] = key + "=" + value
	}
	return strings.Join(parts, "&")
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		if first, _, ok := strings.Cut(forwardedFor, ","); ok {
			return strings.TrimSpace(first)
		}
		return forwardedFor
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func newProxyRequestID() string {
	var randomBytes [8]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + hex.EncodeToString(randomBytes[:])
}

type logResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
	totalBytes int64
	truncated  bool
	limit      int
}

type readCloser struct {
	io.Reader
	io.Closer
}

func newLogResponseRecorder(w http.ResponseWriter, limit int) *logResponseRecorder {
	if limit <= 0 {
		limit = 64 * 1024
	}
	return &logResponseRecorder{
		ResponseWriter: w,
		limit:          limit,
	}
}

func (w *logResponseRecorder) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *logResponseRecorder) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	w.capture(data)
	return w.ResponseWriter.Write(data)
}

func (w *logResponseRecorder) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func (w *logResponseRecorder) BodySnapshot() service.BodySnapshot {
	return service.BodySnapshot{
		Text:      strings.ToValidUTF8(w.body.String(), "?"),
		Bytes:     w.totalBytes,
		Truncated: w.truncated,
	}
}

func (w *logResponseRecorder) capture(data []byte) {
	w.totalBytes += int64(len(data))
	if len(data) == 0 || w.truncated {
		return
	}

	remaining := w.limit - w.body.Len()
	if remaining <= 0 {
		w.truncated = true
		return
	}
	if len(data) > remaining {
		w.body.Write(data[:remaining])
		w.truncated = true
		return
	}
	w.body.Write(data)
}
