package admin

import (
	"fmt"
	"strings"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/types"

	"gorm.io/gorm"
)

const requestLogListOrder = "deleted_at ASC, id DESC"

func applyRequestLogStatusFilter(query *gorm.DB, status string) *gorm.DB {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "success":
		return query.Where("status_code >= ? AND status_code < ?", 200, 400)
	case "error", "failed", "failure":
		return query.Where("status_code = 0 OR status_code < ? OR status_code >= ?", 200, 400)
	default:
		return query
	}
}

func applyRequestLogKeywordFilter(query *gorm.DB, keyword string, columns ...string) *gorm.DB {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" || len(columns) == 0 {
		return query
	}

	conditions := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns))
	pattern := "%" + escapeLikeKeyword(keyword) + "%"
	for _, column := range columns {
		if column = strings.TrimSpace(column); column == "" {
			continue
		}
		// column 仅由调用方传入固定表字段，用户输入只进入参数绑定。
		conditions = append(conditions, fmt.Sprintf(`%s ILIKE ? ESCAPE '\'`, column))
		args = append(args, pattern)
	}
	if len(conditions) == 0 {
		return query
	}
	return query.Where("("+strings.Join(conditions, " OR ")+")", args...)
}

func escapeLikeKeyword(keyword string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(keyword)
}

// countRequestLogs 返回当前筛选条件下的真实日志总数。
// 概览卡片直接展示该值，避免超过一页时固定显示“至少 21”。
func countRequestLogs(query *gorm.DB) (int64, error) {
	var total int64
	err := query.Session(&gorm.Session{}).Count(&total).Error
	return total, err
}

func proxyAccessLogItem(record model.ProxyAccessLog) types.AdminProxyAccessLogItem {
	return types.AdminProxyAccessLogItem{
		Id:                    int64(record.ID),
		RequestId:             record.RequestID,
		Method:                record.Method,
		Path:                  record.Path,
		Query:                 record.Query,
		RequestUri:            record.RequestURI,
		ClientIp:              record.ClientIP,
		UserAgent:             record.UserAgent,
		StatusCode:            record.StatusCode,
		DurationMs:            record.DurationMs,
		ErrorMessage:          record.ErrorMessage,
		RequestBodyBytes:      record.RequestBodyBytes,
		RequestBodyTruncated:  record.RequestBodyTruncated,
		ResponseBodyBytes:     record.ResponseBodyBytes,
		ResponseBodyTruncated: record.ResponseBodyTruncated,
		CreatedAt:             formatLogTime(record.CreatedAt),
	}
}

func proxyAccessLogDetail(record model.ProxyAccessLog) *types.AdminProxyAccessLogDetailResp {
	return &types.AdminProxyAccessLogDetailResp{
		Id:                    int64(record.ID),
		RequestId:             record.RequestID,
		Method:                record.Method,
		Path:                  record.Path,
		Query:                 record.Query,
		RequestUri:            record.RequestURI,
		ClientIp:              record.ClientIP,
		UserAgent:             record.UserAgent,
		StatusCode:            record.StatusCode,
		DurationMs:            record.DurationMs,
		ErrorMessage:          record.ErrorMessage,
		RequestBody:           record.RequestBody,
		RequestBodyBytes:      record.RequestBodyBytes,
		RequestBodyTruncated:  record.RequestBodyTruncated,
		ResponseBody:          record.ResponseBody,
		ResponseBodyBytes:     record.ResponseBodyBytes,
		ResponseBodyTruncated: record.ResponseBodyTruncated,
		CreatedAt:             formatLogTime(record.CreatedAt),
	}
}

func tmdbRequestLogItem(record model.TmdbRequestLog) types.AdminTmdbRequestLogItem {
	return types.AdminTmdbRequestLogItem{
		Id:                    int64(record.ID),
		RequestId:             record.RequestID,
		Method:                record.Method,
		Path:                  record.Path,
		Url:                   record.URL,
		StatusCode:            record.StatusCode,
		DurationMs:            record.DurationMs,
		ErrorMessage:          record.ErrorMessage,
		RequestBodyBytes:      record.RequestBodyBytes,
		RequestBodyTruncated:  record.RequestBodyTruncated,
		ResponseBodyBytes:     record.ResponseBodyBytes,
		ResponseBodyTruncated: record.ResponseBodyTruncated,
		CreatedAt:             formatLogTime(record.CreatedAt),
	}
}

func tmdbRequestLogDetail(record model.TmdbRequestLog) *types.AdminTmdbRequestLogDetailResp {
	return &types.AdminTmdbRequestLogDetailResp{
		Id:                    int64(record.ID),
		RequestId:             record.RequestID,
		Method:                record.Method,
		Path:                  record.Path,
		Url:                   record.URL,
		StatusCode:            record.StatusCode,
		DurationMs:            record.DurationMs,
		ErrorMessage:          record.ErrorMessage,
		RequestBody:           record.RequestBody,
		RequestBodyBytes:      record.RequestBodyBytes,
		RequestBodyTruncated:  record.RequestBodyTruncated,
		ResponseBody:          record.ResponseBody,
		ResponseBodyBytes:     record.ResponseBodyBytes,
		ResponseBodyTruncated: record.ResponseBodyTruncated,
		CreatedAt:             formatLogTime(record.CreatedAt),
	}
}

func truncateRequestLogTable(db *gorm.DB, modelValue interface{}) error {
	tableName, err := requestLogTableName(db, modelValue)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY", tableName)
	return db.Exec(sql).Error
}

func requestLogTableName(db *gorm.DB, modelValue interface{}) (string, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(modelValue); err != nil {
		return "", err
	}
	return stmt.Table, nil
}
