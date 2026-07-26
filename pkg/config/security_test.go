package config

import (
	"strings"
	"testing"
)

func environment(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestSecurityConfigurationFailsClosed(t *testing.T) {
	for _, values := range []map[string]string{
		{},
		{"OPENREVENUE_ENV": "production"},
		{
			"OPENREVENUE_ENV":         "production",
			"DATABASE_URL_FILE":       "/run/secrets/database",
			"OIDC_CLIENT_SECRET_FILE": "/run/secrets/oidc",
			"S3_CREDENTIALS_FILE":     "/run/secrets/s3",
			"TLS_TERMINATED":          "false",
		},
		{
			"OPENREVENUE_ENV":         "production",
			"DATABASE_URL_FILE":       "/run/secrets/database",
			"OIDC_CLIENT_SECRET_FILE": "/run/secrets/oidc",
			"S3_CREDENTIALS_FILE":     "/run/secrets/s3",
			"TLS_TERMINATED":          "true",
			"S3_SECRET_KEY":           "raw-secret",
		},
	} {
		if _, err := loadSecurity(environment(values)); err == nil {
			t.Fatalf("unsafe configuration was accepted: %#v", values)
		}
	}
}

func TestProductionAcceptsInjectedSecretsAndTLS(t *testing.T) {
	security, err := loadSecurity(environment(map[string]string{
		"OPENREVENUE_ENV":         "production",
		"DATABASE_URL_FILE":       "/run/secrets/database-url",
		"OIDC_CLIENT_SECRET_FILE": "/run/secrets/oidc-client",
		"S3_CREDENTIALS_FILE":     "/run/secrets/s3-credentials",
		"TLS_TERMINATED":          "true",
	}))
	if err != nil || security.Environment != "production" {
		t.Fatalf("safe production configuration rejected: %#v %v", security, err)
	}
}

func TestDevelopmentMustStillBeExplicit(t *testing.T) {
	security, err := loadSecurity(environment(map[string]string{
		"OPENREVENUE_ENV": "development",
	}))
	if err != nil || !strings.EqualFold(security.Environment, "development") {
		t.Fatalf("explicit development configuration rejected: %#v %v", security, err)
	}
}
