package config

import (
	"os"
)

type Config struct {
	ProjectID              string
	OAuthClientID          string
	OAuthClientSecret      string
	Port                   string
	DBUser                 string
	DBPass                 string
	DBName                 string
	InstanceConnectionName string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		ProjectID:              os.Getenv("GOOGLE_CLOUD_PROJECT"),
		OAuthClientID:          os.Getenv("OAUTH_CLIENT_ID"),
		OAuthClientSecret:      os.Getenv("OAUTH_CLIENT_SECRET"),
		Port:                   port,
		DBUser:                 os.Getenv("DB_USER"),
		DBPass:                 os.Getenv("DB_PASS"),
		DBName:                 os.Getenv("DB_NAME"),
		InstanceConnectionName: os.Getenv("INSTANCE_CONNECTION_NAME"),
	}
}
