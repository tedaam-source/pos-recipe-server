package config

import (
	"os"
	"strings"
)

type Config struct {
	ProjectID              string
	Port                   string
	DBUser                 string
	DBPass                 string
	DBName                 string
	InstanceConnectionName string
	AdminAllowlist         []string
	WorkerBaseURL          string
	AppEnv                 string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default to 8081 for admin locally to avoid conflict
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "production"
	}

	allowlistStr := os.Getenv("ADMIN_ALLOWLIST")
	var allowlist []string
	if allowlistStr != "" {
		for _, s := range strings.Split(allowlistStr, ",") {
			allowlist = append(allowlist, strings.TrimSpace(s))
		}
	}

	return &Config{
		ProjectID:              os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Port:                   port,
		DBUser:                 os.Getenv("DB_USER"),
		DBPass:                 getEnv("DB_PASSWORD", "DB_PASS"), // Support both
		DBName:                 os.Getenv("DB_NAME"),
		InstanceConnectionName: getEnv("DB_INSTANCE_CONNECTION_NAME", "INSTANCE_CONNECTION_NAME"), // Support both
		AdminAllowlist:         allowlist,
		WorkerBaseURL:          os.Getenv("WORKER_BASE_URL"),
		AppEnv:                 appEnv,
	}
}

func getEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
