package service

import "testing"

func TestParseProxyMediaTarget(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		wantMediaType string
		wantTmdbID    int
	}{
		{
			name:          "api v3 movie detail",
			path:          "/api/v3/movie/550",
			wantMediaType: "movie",
			wantTmdbID:    550,
		},
		{
			name:          "v3 tv nested path",
			path:          "/v3/tv/1399/season/1",
			wantMediaType: "tv",
			wantTmdbID:    1399,
		},
		{
			name:          "tmdb path without local prefix",
			path:          "/movie/299536/images",
			wantMediaType: "movie",
			wantTmdbID:    299536,
		},
		{
			name:          "local negative id",
			path:          "/3/movie/-1",
			wantMediaType: "movie",
			wantTmdbID:    -1,
		},
		{
			name: "non media path",
			path: "/3/search/movie",
		},
		{
			name: "invalid id",
			path: "/3/tv/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMediaType, gotTmdbID := parseProxyMediaTarget(tt.path)
			if gotMediaType != tt.wantMediaType || gotTmdbID != tt.wantTmdbID {
				t.Fatalf("parseProxyMediaTarget(%q) = (%q, %d), want (%q, %d)", tt.path, gotMediaType, gotTmdbID, tt.wantMediaType, tt.wantTmdbID)
			}
		})
	}
}
