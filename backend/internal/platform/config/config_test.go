package config

import (
	"strings"
	"testing"
)

func validEnvironment(t *testing.T) {
	t.Helper()
	for name, value := range map[string]string{
		"MYSQL_DSN":     "app:secret@tcp(localhost:3306)/just_for_you?parseTime=true",
		"JWT_SECRET":    strings.Repeat("s", 32),
		"WECHAT_APP_ID": "wx-test", "WECHAT_APP_SECRET": "wechat-secret",
		"COS_BUCKET_URL": "https://bucket.cos.example.com", "COS_SECRET_ID": "cos-id", "COS_SECRET_KEY": "cos-secret",
	} {
		t.Setenv(name, value)
	}
}

func TestLoadReturnsErrorWhenRequiredSecretIsMissing(t *testing.T) {
	validEnvironment(t)
	for _, name := range []string{"MYSQL_DSN", "JWT_SECRET", "WECHAT_APP_SECRET", "COS_SECRET_KEY"} {
		t.Run(name, func(t *testing.T) {
			validEnvironment(t)
			t.Setenv(name, "")
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), name) {
				t.Fatalf("Load() error = %v, want error naming %s", err, name)
			}
		})
	}
}

func TestLoadRejectsShortJWTSecret(t *testing.T) {
	validEnvironment(t)
	t.Setenv("JWT_SECRET", "too-short")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "at least 32 bytes") {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestLoadEnablesSoloTestOnlyWhenExplicitlyTrue(t *testing.T) {
	validEnvironment(t)
	t.Setenv("ENABLE_SOLO_TEST", "true")
	cfg, err := Load()
	if err != nil || !cfg.EnableSoloTest {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
	t.Setenv("ENABLE_SOLO_TEST", "")
	cfg, err = Load()
	if err != nil || cfg.EnableSoloTest {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
}

func TestLoadTreatsAIConfigurationAsOptionalGroup(t *testing.T) {
	validEnvironment(t)
	t.Setenv("AI_BASE_URL", "")
	t.Setenv("AI_API_KEY", "")
	t.Setenv("AI_MODEL", "")
	cfg, err := Load()
	if err != nil || cfg.AIBaseURL != "" {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
	t.Setenv("AI_BASE_URL", "https://ai.example/v1")
	t.Setenv("AI_API_KEY", "test-key")
	t.Setenv("AI_MODEL", "small")
	cfg, err = Load()
	if err != nil || cfg.AIModel != "small" {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
}
