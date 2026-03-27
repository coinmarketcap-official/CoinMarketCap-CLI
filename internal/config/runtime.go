package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type AuthSource string

const (
	AuthSourceNone   AuthSource = "none"
	AuthSourceEnv    AuthSource = "env"
	AuthSourceConfig AuthSource = "config"
)

type RuntimeConfig struct {
	Effective     *Config
	AuthSource    AuthSource
	Configured    bool
	Warning       string
	ConfigWarning string
	ConfigPath    string
}

func LoadRuntime() (*RuntimeConfig, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	runtime := &RuntimeConfig{
		AuthSource: AuthSourceNone,
		ConfigPath: path,
	}

	fileCfg, filePresent, warning, err := loadFileConfig(path)
	if err != nil {
		return nil, err
	}
	runtime.Configured = filePresent && fileCfg != nil && strings.TrimSpace(fileCfg.APIKey) != ""
	runtime.Warning = warning
	runtime.ConfigWarning = warning

	if envCfg, ok := loadEnvConfig(); ok {
		runtime.AuthSource = AuthSourceEnv
		runtime.Effective = envCfg
		return runtime, nil
	}

	if filePresent && fileCfg != nil && strings.TrimSpace(fileCfg.APIKey) != "" {
		runtime.AuthSource = AuthSourceConfig
		runtime.Effective = fileCfg
		return runtime, nil
	}

	return runtime, nil
}

func loadFileConfig(path string) (*Config, bool, string, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false, "", nil
		}
		return nil, false, "", fmt.Errorf("failed to stat config: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, false, "", fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, true, "", fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Tier = strings.ToLower(strings.TrimSpace(cfg.Tier))

	if cfg.APIKey == "" {
		return &cfg, true, "config missing api_key", nil
	}
	if cfg.Tier != "" && !IsValidTier(cfg.Tier) {
		invalidTier := cfg.Tier
		cfg.Tier = ""
		return &cfg, true, fmt.Sprintf("invalid persisted tier %q", invalidTier), nil
	}
	return &cfg, true, "", nil
}

func loadEnvConfig() (*Config, bool) {
	key := strings.TrimSpace(os.Getenv("CMC_API_KEY"))
	tier := strings.ToLower(strings.TrimSpace(os.Getenv("CMC_API_TIER")))
	if key == "" {
		return nil, false
	}
	if tier != "" && !IsValidTier(tier) {
		tier = ""
	}
	return &Config{APIKey: key, Tier: tier}, true
}
