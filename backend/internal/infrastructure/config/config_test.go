package config

import (
	"strings"
	"testing"
)

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("MYSQL_DSN", "app:secret@tcp(localhost:3306)/just_for_you?parseTime=true")
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("WECHAT_APP_ID", "wx-test")
	t.Setenv("WECHAT_APP_SECRET", "wechat-secret")
	t.Setenv("COS_BUCKET_URL", "https://bucket.cos.example.com")
	t.Setenv("COS_SECRET_ID", "cos-id")
	t.Setenv("COS_SECRET_KEY", "cos-secret")
}

func TestLoadRejectsMissingMySQLDSN(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("MYSQL_DSN", "")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "MYSQL_DSN") {
		t.Fatalf("Load() error = %v, want error naming MYSQL_DSN", err)
	}
}

func TestLoadRejectsShortJWTSecret(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("JWT_SECRET", "too-short")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "at least 32 bytes") {
		t.Fatalf("Load() error = %v, want minimum-length error", err)
	}
}

func TestLoadEnablesSoloTestOnlyWhenExplicitlyTrue(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("ENABLE_SOLO_TEST", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EnableSoloTest {
		t.Fatal("EnableSoloTest=false, want true")
	}

	t.Setenv("ENABLE_SOLO_TEST", "")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableSoloTest {
		t.Fatal("EnableSoloTest=true without explicit opt-in")
	}
}

func TestAIConfigurationIsOptionalAndLoadedAsAGroup(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("AI_BASE_URL", "")
	t.Setenv("AI_API_KEY", "")
	t.Setenv("AI_MODEL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AIBaseURL != "" {
		t.Fatalf("cfg=%+v", cfg)
	}
	t.Setenv("AI_BASE_URL", "https://ai.example/v1")
	t.Setenv("AI_API_KEY", "test-key")
	t.Setenv("AI_MODEL", "small")
	cfg, err = Load()
	if err != nil || cfg.AIModel != "small" {
		t.Fatalf("cfg=%+v err=%v", cfg, err)
	}
}
