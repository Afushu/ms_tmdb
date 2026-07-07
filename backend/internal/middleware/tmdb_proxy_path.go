package middleware

import "strings"

const tmdbProxyPrefix = "/api/tmdb"

func isTmdbProxyPath(path string) bool {
	return path == tmdbProxyPrefix || strings.HasPrefix(path, tmdbProxyPrefix+"/")
}

func resolveTmdbPath(path string) (string, bool) {
	if path == tmdbProxyPrefix {
		return "", true
	}
	if strings.HasPrefix(path, tmdbProxyPrefix+"/") {
		return strings.TrimPrefix(path, tmdbProxyPrefix), true
	}
	return "", false
}
