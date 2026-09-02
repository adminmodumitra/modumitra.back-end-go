package config

import (
	"os"
	"strings"
)

func resolveCabinetPath() string {
	if p := strings.TrimSpace(os.Getenv("CABINET_CONFIG")); p != "" {
		return p
	}
	return findRel(cabinetFileForEnv(appEnv()))
}

func appEnv() string {
	if v := strings.TrimSpace(os.Getenv("APP_ENV")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("CABINET_ENV")); v != "" {
		return v
	}
	return "local"
}

func cabinetFileForEnv(env string) string {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "dev", "development":
		return "configs/cabinet.dev.json"
	case "production", "prod":
		return "configs/cabinet.production.json"
	default:
		return defaultCabinetPath
	}
}
