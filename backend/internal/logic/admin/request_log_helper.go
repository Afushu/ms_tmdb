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

func requestLogPageLimit(pageSize int) int {
	return pageSize + 1
}

func requestLogVisibleCount(pageSize int, loaded int) int {
	if loaded > pageSize {
		return pageSize
	}
	return loaded
}

// requestLogWindowTotal 返回当前分页窗口能证明的总量下界。
// 请求日志表增长快，精确 COUNT 会随表大小线性变慢；多取一条即可判断是否允许翻到下一页。
func requestLogWindowTotal(page int, pageSize int, loaded int) int64 {
	offset := (page - 1) * pageSize
	visible := requestLogVisibleCount(pageSize, loaded)
	total := int64(offset + visible)
	if loaded > pageSize {
		return total + 1
	}
	return total
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
