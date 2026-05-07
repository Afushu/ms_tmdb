package tmdbclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultHTTPTimeout       = 15 * time.Second
	defaultRequestLogTimeout = 3 * time.Second
)

// APIError 表示 TMDB 上游返回的非 200 响应，保留状态码便于上层按场景映射提示。
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("TMDB 返回错误状态码 %d: %s", e.StatusCode, e.Body)
}

// StatusMessage 读取 TMDB 错误响应中的 status_message，body 非 JSON 时返回空字符串。
func (e *APIError) StatusMessage() string {
	var payload struct {
		StatusMessage string `json:"status_message"`
	}
	if err := json.Unmarshal([]byte(e.Body), &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.StatusMessage)
}

// Client TMDB API 客户端
type Client struct {
	apiKey     string
	baseURL    string
	language   string
	httpClient *http.Client
	proxyURL   string
	mu         sync.RWMutex

	// 简单令牌桶限流
	rateLimiter chan struct{}

	requestLogger RequestLogger
}

// NewClient 创建 TMDB 客户端
func NewClient(apiKey, baseURL, defaultLanguage string, rateLimit int, proxyURL string) *Client {
	if rateLimit <= 0 {
		logx.Errorf("TMDB RateLimit 配置无效(%d)，已回退为 40", rateLimit)
		rateLimit = 40
	}

	c := &Client{
		apiKey:   apiKey,
		baseURL:  baseURL,
		language: defaultLanguage,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		rateLimiter: make(chan struct{}, rateLimit),
	}

	if err := c.SetProxy(proxyURL); err != nil {
		logx.Errorf("初始化 TMDB 代理失败，已降级为直连: %v", err)
	}

	// 填充令牌桶
	for i := 0; i < rateLimit; i++ {
		c.rateLimiter <- struct{}{}
	}

	// 定时补充令牌
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rateLimit))
		defer ticker.Stop()
		for range ticker.C {
			select {
			case c.rateLimiter <- struct{}{}:
			default:
			}
		}
	}()

	return c
}

// GetProxy 返回当前代理地址
func (c *Client) GetProxy() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.proxyURL
}

// SetProxy 设置 TMDB 请求代理地址，空字符串表示关闭自定义代理
func (c *Client) SetProxy(proxyURL string) error {
	trimmed := strings.TrimSpace(proxyURL)
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	if trimmed != "" {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return fmt.Errorf("代理地址格式不正确: %w", err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("代理地址必须包含协议和主机，例如 http://127.0.0.1:7890")
		}
		transport.Proxy = http.ProxyURL(parsed)
	}

	httpClient := &http.Client{
		Timeout:   defaultHTTPTimeout,
		Transport: transport,
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.httpClient = httpClient
	c.proxyURL = trimmed
	return nil
}

// RequestOption 请求选项
type RequestOption struct {
	Context          context.Context
	RequestID        string
	Language         string
	Page             int
	Region           string
	AppendToResponse string
	ExtraParams      map[string]string
}

// RequestLogEntry 描述一次真实 TMDB 上游请求，供业务层选择性落库。
type RequestLogEntry struct {
	RequestID string
	Method    string
	Path      string
	URL       string

	StatusCode   int
	DurationMs   int64
	ErrorMessage string

	RequestBody  []byte
	ResponseBody []byte
}

// RequestLogger 是 TMDB 请求日志回调。回调失败不能影响原始请求。
type RequestLogger func(ctx context.Context, entry RequestLogEntry)

// SetRequestLogger 设置真实 TMDB 请求日志回调。
func (c *Client) SetRequestLogger(logger RequestLogger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestLogger = logger
}

// Get 发送 GET 请求到 TMDB API
func (c *Client) Get(path string, opts *RequestOption) (json.RawMessage, error) {
	ctx := context.Background()
	if opts != nil && opts.Context != nil {
		ctx = opts.Context
	}

	// 等待限流令牌时也要响应请求取消，避免客户端断开后继续阻塞。
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.rateLimiter:
	}

	reqURL, err := c.buildURL(path, opts)
	if err != nil {
		return nil, fmt.Errorf("构建请求 URL 失败: %w", err)
	}

	logx.Debugf("TMDB 请求: %s", maskSensitiveQuery(reqURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	c.mu.RLock()
	httpClient := c.httpClient
	c.mu.RUnlock()

	startedAt := time.Now()
	logEntry := RequestLogEntry{
		Method:    http.MethodGet,
		Path:      path,
		URL:       maskSensitiveQuery(reqURL),
		RequestID: requestIDFromOption(opts),
	}
	defer func() {
		logEntry.DurationMs = time.Since(startedAt).Milliseconds()
		c.writeRequestLog(ctx, logEntry)
	}()

	resp, err := httpClient.Do(req)
	if err != nil {
		logEntry.ErrorMessage = err.Error()
		return nil, fmt.Errorf("TMDB 请求失败: %w", err)
	}
	defer resp.Body.Close()
	logEntry.StatusCode = resp.StatusCode

	body, err := io.ReadAll(resp.Body)
	logEntry.ResponseBody = body
	if err != nil {
		logEntry.ErrorMessage = err.Error()
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
		logEntry.ErrorMessage = apiErr.Error()
		return nil, apiErr
	}

	return json.RawMessage(body), nil
}

func (c *Client) writeRequestLog(ctx context.Context, entry RequestLogEntry) {
	c.mu.RLock()
	logger := c.requestLogger
	c.mu.RUnlock()
	if logger == nil {
		return
	}

	entry = cloneRequestLogEntry(entry)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				logx.Errorf("TMDB 请求日志回调异常: %v", recovered)
			}
		}()

		logCtx, cancel := context.WithTimeout(detachedContext(ctx), defaultRequestLogTimeout)
		defer cancel()
		logger(logCtx, entry)
	}()
}

func requestIDFromOption(opts *RequestOption) string {
	if opts == nil {
		return ""
	}
	return strings.TrimSpace(opts.RequestID)
}

func cloneRequestLogEntry(entry RequestLogEntry) RequestLogEntry {
	entry.RequestBody = append([]byte(nil), entry.RequestBody...)
	entry.ResponseBody = append([]byte(nil), entry.ResponseBody...)
	return entry
}

func detachedContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}

// Request 通用请求方法（Get 的别名），供中间件透传路径使用
func (c *Client) Request(path string, opts *RequestOption) (json.RawMessage, error) {
	return c.Get(path, opts)
}

// buildURL 构建完整的 TMDB API URL
func (c *Client) buildURL(path string, opts *RequestOption) (string, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("api_key", c.apiKey)

	// 语言参数
	lang := c.language
	if opts != nil && opts.Language != "" {
		lang = opts.Language
	}
	q.Set("language", lang)

	if opts != nil {
		if opts.Page > 0 {
			q.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.Region != "" {
			q.Set("region", opts.Region)
		}
		if opts.AppendToResponse != "" {
			q.Set("append_to_response", opts.AppendToResponse)
		}
		for k, v := range opts.ExtraParams {
			q.Set(k, v)
		}
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func maskSensitiveQuery(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := u.Query()
	for key := range q {
		if strings.EqualFold(key, "api_key") {
			q.Set(key, "***")
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
