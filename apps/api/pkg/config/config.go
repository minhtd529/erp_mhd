package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
	JWT      JWTConfig
	TwoFA    TwoFAConfig
	HRM      HRMConfig
}

type AppConfig struct {
	Env  string `mapstructure:"ENV"`
	Port string `mapstructure:"PORT"`
	Name string `mapstructure:"APP_NAME"`
}

type DatabaseConfig struct {
	URL         string `mapstructure:"DATABASE_URL"`
	MaxConns    int    `mapstructure:"DB_MAX_CONNS"`
	MinConns    int    `mapstructure:"DB_MIN_CONNS"`
	MaxIdleTime string `mapstructure:"DB_MAX_IDLE_TIME"`
}

type RedisConfig struct {
	URL string `mapstructure:"REDIS_URL"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"MINIO_ENDPOINT"`
	AccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	SecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	Bucket    string `mapstructure:"MINIO_BUCKET"`
	UseSSL    bool   `mapstructure:"MINIO_USE_SSL"`
}

type JWTConfig struct {
	Secret          string `mapstructure:"JWT_SECRET"`
	AccessTokenTTL  int    `mapstructure:"JWT_ACCESS_TTL_MINUTES"`
	RefreshTokenTTL int    `mapstructure:"JWT_REFRESH_TTL_DAYS"`
}

type HRMConfig struct {
	BankEncryptionKey string `mapstructure:"HRM_BANK_ENCRYPTION_KEY"` // 64-char hex = 32 raw bytes
	EncryptionKey     string `mapstructure:"HRM_ENCRYPTION_KEY"`      // base64-encoded 32 bytes for AES-256-GCM PII encryption
}

type TwoFAConfig struct {
	EncryptionKey     string `mapstructure:"TOTP_ENCRYPTION_KEY"` // 64-char hex = 32 raw bytes
	ChallengeTTLSecs  int    `mapstructure:"TOTP_CHALLENGE_TTL_SECS"`
	MaxAttempts       int    `mapstructure:"TOTP_MAX_ATTEMPTS"`
	TrustDeviceDays   int    `mapstructure:"TOTP_TRUST_DEVICE_DAYS"`
	MaxTrustedDevices int    `mapstructure:"TOTP_MAX_TRUSTED_DEVICES"`
	Issuer            string `mapstructure:"TOTP_ISSUER"`
}

func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("ENV", "development")
	v.SetDefault("PORT", "8080")
	v.SetDefault("APP_NAME", "erp-audit-api")
	v.SetDefault("DATABASE_URL", "postgres://erp:erp@localhost:5432/erp_audit?sslmode=disable")
	v.SetDefault("DB_MAX_CONNS", 25)
	v.SetDefault("DB_MIN_CONNS", 5)
	v.SetDefault("DB_MAX_IDLE_TIME", "15m")
	v.SetDefault("REDIS_URL", "redis://localhost:6379/0")
	v.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	v.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	v.SetDefault("MINIO_SECRET_KEY", "minioadmin")
	v.SetDefault("MINIO_BUCKET", "erp-audit")
	v.SetDefault("MINIO_USE_SSL", false)
	v.SetDefault("JWT_SECRET", "change-me-in-production")
	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 15)
	v.SetDefault("JWT_REFRESH_TTL_DAYS", 7)

	// HRM bank encryption key — must be overridden in production
	v.SetDefault("HRM_BANK_ENCRYPTION_KEY", "0000000000000000000000000000000000000000000000000000000000000000")
	// HRM PII encryption key (base64-encoded 32 bytes) — MUST be set in production
	v.SetDefault("HRM_ENCRYPTION_KEY", "")

	// 2FA defaults — TOTP_ENCRYPTION_KEY must be overridden in production
	v.SetDefault("TOTP_ENCRYPTION_KEY", "0000000000000000000000000000000000000000000000000000000000000000")
	v.SetDefault("TOTP_CHALLENGE_TTL_SECS", 300) // 5 minutes
	v.SetDefault("TOTP_MAX_ATTEMPTS", 5)
	v.SetDefault("TOTP_TRUST_DEVICE_DAYS", 30)
	v.SetDefault("TOTP_MAX_TRUSTED_DEVICES", 5)
	v.SetDefault("TOTP_ISSUER", "ERP Audit")

	// Read from .env file if present
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // Ignore error if .env not found

	// Override with actual env vars
	v.AutomaticEnv()

	cfg := &Config{}
	cfg.App = AppConfig{
		Env:  v.GetString("ENV"),
		Port: v.GetString("PORT"),
		Name: v.GetString("APP_NAME"),
	}
	cfg.Database = DatabaseConfig{
		URL:         v.GetString("DATABASE_URL"),
		MaxConns:    v.GetInt("DB_MAX_CONNS"),
		MinConns:    v.GetInt("DB_MIN_CONNS"),
		MaxIdleTime: v.GetString("DB_MAX_IDLE_TIME"),
	}
	cfg.Redis = RedisConfig{
		URL: v.GetString("REDIS_URL"),
	}
	cfg.MinIO = MinIOConfig{
		Endpoint:  v.GetString("MINIO_ENDPOINT"),
		AccessKey: v.GetString("MINIO_ACCESS_KEY"),
		SecretKey: v.GetString("MINIO_SECRET_KEY"),
		Bucket:    v.GetString("MINIO_BUCKET"),
		UseSSL:    v.GetBool("MINIO_USE_SSL"),
	}
	cfg.JWT = JWTConfig{
		Secret:          v.GetString("JWT_SECRET"),
		AccessTokenTTL:  v.GetInt("JWT_ACCESS_TTL_MINUTES"),
		RefreshTokenTTL: v.GetInt("JWT_REFRESH_TTL_DAYS"),
	}
	cfg.TwoFA = TwoFAConfig{
		EncryptionKey:     v.GetString("TOTP_ENCRYPTION_KEY"),
		ChallengeTTLSecs:  v.GetInt("TOTP_CHALLENGE_TTL_SECS"),
		MaxAttempts:       v.GetInt("TOTP_MAX_ATTEMPTS"),
		TrustDeviceDays:   v.GetInt("TOTP_TRUST_DEVICE_DAYS"),
		MaxTrustedDevices: v.GetInt("TOTP_MAX_TRUSTED_DEVICES"),
		Issuer:            v.GetString("TOTP_ISSUER"),
	}
	cfg.HRM = HRMConfig{
		BankEncryptionKey: v.GetString("HRM_BANK_ENCRYPTION_KEY"),
		EncryptionKey:     v.GetString("HRM_ENCRYPTION_KEY"),
	}

	return cfg, nil
}
