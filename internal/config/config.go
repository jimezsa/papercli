package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	configFile = "config.json"
	cacheFile  = "cache.db"
)

type Config struct {
	DefaultProvider string `json:"default_provider"`
	DefaultLimit    int    `json:"default_limit"`
	SemanticAPIKey  string `json:"semantic_api_key,omitempty"`
	SerpAPIKey      string `json:"serpapi_key,omitempty"`
}

func Default() Config {
	return Config{
		DefaultProvider: "all",
		DefaultLimit:    20,
	}
}

func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, "papercli"), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFile), nil
}

func CachePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, cacheFile), nil
}

func EnsureFile() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, configFile)
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat config %q: %w", path, err)
	}
	if err := writeConfig(path, Default()); err != nil {
		return "", err
	}
	return path, nil
}

func InitFile(force bool) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, configFile)
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return "", fmt.Errorf("config already exists at %q (use --force to overwrite)", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat config %q: %w", path, err)
		}
	}

	if err := writeConfig(path, Default()); err != nil {
		return "", err
	}
	return path, nil
}

func Load() (Config, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		if !errors.Is(readErr, os.ErrNotExist) {
			return cfg, fmt.Errorf("read config %q: %w", path, readErr)
		}
		applyEnvOverrides(&cfg)
		return cfg, nil
	}

	if len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("decode config %q: %w", path, err)
		}
	}

	applyEnvOverrides(&cfg)
	if cfg.DefaultProvider == "" {
		cfg.DefaultProvider = "all"
	}
	if cfg.DefaultLimit <= 0 {
		cfg.DefaultLimit = 20
	}
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("PAPERCLI_SEMANTIC_API_KEY"); v != "" {
		cfg.SemanticAPIKey = v
	}
	if v := os.Getenv("PAPERCLI_SERPAPI_KEY"); v != "" {
		cfg.SerpAPIKey = v
	}
	if v := os.Getenv("PAPERCLI_DEFAULT_PROVIDER"); v != "" {
		cfg.DefaultProvider = strings.ToLower(strings.TrimSpace(v))
	}
	if v := os.Getenv("PAPERCLI_DEFAULT_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.DefaultLimit = n
		}
	}
}

func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	return nil
}

func writeConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
