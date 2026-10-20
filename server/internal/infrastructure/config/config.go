package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr     string
	MySQLDSN     string
	JWTSecret    string
	WeChatAppID  string
	WeChatSecret string
	COSBucketURL string
	COSSecretID  string
	COSSecretKey string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr: os.Getenv("HTTP_ADDR"), MySQLDSN: os.Getenv("MYSQL_DSN"),
		JWTSecret: os.Getenv("JWT_SECRET"), WeChatAppID: os.Getenv("WECHAT_APP_ID"),
		WeChatSecret: os.Getenv("WECHAT_APP_SECRET"), COSBucketURL: os.Getenv("COS_BUCKET_URL"),
		COSSecretID: os.Getenv("COS_SECRET_ID"), COSSecretKey: os.Getenv("COS_SECRET_KEY"),
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	for name, value := range map[string]string{
		"MYSQL_DSN": cfg.MySQLDSN, "JWT_SECRET": cfg.JWTSecret,
		"WECHAT_APP_ID": cfg.WeChatAppID, "WECHAT_APP_SECRET": cfg.WeChatSecret,
		"COS_BUCKET_URL": cfg.COSBucketURL, "COS_SECRET_ID": cfg.COSSecretID,
		"COS_SECRET_KEY": cfg.COSSecretKey,
	} {
		if value == "" {
			return Config{}, fmt.Errorf("%s is required", name)
		}
	}
	if len([]byte(cfg.JWTSecret)) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	return cfg, nil
}
