// Package config validates fail-closed runtime configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Security struct {
	Environment string
}

func LoadSecurity() (Security, error) {
	return loadSecurity(os.Getenv)
}

func loadSecurity(getenv func(string) string) (Security, error) {
	environment := strings.ToLower(strings.TrimSpace(getenv("OPENREVENUE_ENV")))
	switch environment {
	case "development", "test":
		return Security{Environment: environment}, nil
	case "production":
	default:
		return Security{}, errors.New("OPENREVENUE_ENV must explicitly be development, test, or production")
	}

	for _, name := range []string{
		"DATABASE_URL_FILE",
		"OIDC_CLIENT_SECRET_FILE",
		"S3_CREDENTIALS_FILE",
	} {
		value := strings.TrimSpace(getenv(name))
		if value == "" || !strings.HasPrefix(value, "/") {
			return Security{}, fmt.Errorf("%s must reference an absolute injected secret file", name)
		}
	}
	if getenv("TLS_TERMINATED") != "true" {
		return Security{}, errors.New("production requires TLS_TERMINATED=true")
	}
	for _, name := range []string{
		"DATABASE_URL", "OIDC_CLIENT_SECRET", "S3_SECRET_KEY", "SMTP_PASSWORD",
	} {
		if strings.TrimSpace(getenv(name)) != "" {
			return Security{}, fmt.Errorf("%s must not contain a raw production secret; use its _FILE setting", name)
		}
	}
	return Security{Environment: environment}, nil
}
