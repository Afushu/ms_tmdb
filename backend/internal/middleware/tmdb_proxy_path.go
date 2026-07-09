package middleware

import "ms_tmdb/pkg/tmdbpath"

func isTmdbProxyPath(path string) bool {
	return tmdbpath.IsProxyPath(path)
}

func resolveTmdbPath(path string) (string, bool) {
	return tmdbpath.Resolve(path)
}
