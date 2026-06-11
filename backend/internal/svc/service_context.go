package svc

import (
	"context"
	"ms_tmdb/config"
	"ms_tmdb/internal/logic/proxy"
	"ms_tmdb/internal/model"
	"ms_tmdb/internal/service"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	StartupTimeout int64
	DB             *gorm.DB
	TmdbClient     *tmdbclient.Client
	ProxyService   *proxy.ProxyService
	LogService     *service.RequestLogService
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 PostgreSQL 连接
	db, err := gorm.Open(postgres.Open(c.Postgres.DSN()), &gorm.Config{})
	if err != nil {
		logx.Must(err)
	}

	// 仅在模型、索引或历史清理 SQL 变化时执行启动迁移，避免每次启动都触发慢元数据查询。
	if err := model.RunStartupMigrations(db); err != nil {
		logx.Must(err)
	}

	// 初始化 TMDB 客户端
	client := tmdbclient.NewClient(
		c.Tmdb.ApiKey,
		c.Tmdb.BaseURL,
		c.Tmdb.DefaultLanguage,
		c.Tmdb.RateLimit,
		c.Tmdb.ProxyURL,
	)
	client.SetTimeoutMillis(c.Timeout)
	logService := service.NewRequestLogService(db, c.Tmdb.Log)
	client.SetRequestLogger(func(ctx context.Context, entry tmdbclient.RequestLogEntry) {
		err := logService.WriteTmdbRequest(ctx, service.TmdbRequestEntry{
			RequestID: entry.RequestID,
			Method:    entry.Method,
			Path:      entry.Path,
			URL:       entry.URL,

			StatusCode:   entry.StatusCode,
			DurationMs:   entry.DurationMs,
			ErrorMessage: entry.ErrorMessage,

			RequestBody:  logService.CaptureBody(entry.RequestBody),
			ResponseBody: logService.CaptureBody(entry.ResponseBody),
		})
		if err != nil {
			logx.Errorf("写入 TMDB 请求日志失败: %v", err)
		}
	})

	return &ServiceContext{
		Config:         c,
		StartupTimeout: c.Timeout,
		DB:             db,
		TmdbClient:     client,
		ProxyService:   proxy.NewProxyService(db, client, c.Tmdb.DefaultLanguage, c.Tmdb.LocalWriteEnabled),
		LogService:     logService,
	}
}
