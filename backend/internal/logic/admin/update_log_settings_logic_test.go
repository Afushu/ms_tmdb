package admin

import (
	"strings"
	"testing"
)

func TestValidateLogRetentionDays(t *testing.T) {
	if err := validateLogRetentionDays(7); err != nil {
		t.Fatalf("validateLogRetentionDays(7) error = %v", err)
	}
	if err := validateLogRetentionDays(0); err == nil {
		t.Fatal("validateLogRetentionDays(0) expected error")
	}
	if err := validateLogRetentionDays(366); err == nil {
		t.Fatal("validateLogRetentionDays(366) expected error")
	}
}

func TestApplyLogSettingsConfigValuesInsertsLogBlock(t *testing.T) {
	lines := []string{
		"Name: ms-tmdb-api",
		"Tmdb:",
		"  ApiKey: \"x\"",
		"  AutoSync:",
		"    Enabled: true",
	}

	got, err := applyLogSettingsConfigValues(lines, 7, 1048576)
	if err != nil {
		t.Fatalf("applyLogSettingsConfigValues error = %v", err)
	}

	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "  Log:") {
		t.Fatalf("missing Log block:\n%s", joined)
	}
	if !strings.Contains(joined, "    RetentionDays: 7") {
		t.Fatalf("missing RetentionDays:\n%s", joined)
	}
	if !strings.Contains(joined, "    BodyLimitBytes: 1048576") {
		t.Fatalf("missing BodyLimitBytes:\n%s", joined)
	}
}

func TestApplyLogSettingsConfigValuesUpdatesExistingLogBlock(t *testing.T) {
	lines := []string{
		"Tmdb:",
		"  ApiKey: \"x\"",
		"  Log:",
		"    RetentionDays: 7",
		"    BodyLimitBytes: 2048",
		"  AutoSync:",
		"    Enabled: true",
	}

	got, err := applyLogSettingsConfigValues(lines, 21, 0)
	if err != nil {
		t.Fatalf("applyLogSettingsConfigValues error = %v", err)
	}

	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "    RetentionDays: 21") {
		t.Fatalf("RetentionDays not updated:\n%s", joined)
	}
	if !strings.Contains(joined, "    BodyLimitBytes: 2048") {
		t.Fatalf("existing BodyLimitBytes not preserved:\n%s", joined)
	}
	if strings.Count(joined, "  Log:") != 1 {
		t.Fatalf("Log block duplicated:\n%s", joined)
	}
}
