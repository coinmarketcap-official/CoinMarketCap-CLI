package cmd

import (
	"github.com/coinmarketcap/coinmarketcap-cli/internal/api"
	"github.com/coinmarketcap/coinmarketcap-cli/internal/config"
)

// userAgent is the User-Agent header sent with all API requests.
var userAgent = "cmc-cli/" + version

// newAPIClient is the factory used by command handlers to create API clients.
// Tests override this to inject httptest-backed clients.
var newAPIClient = func(cfg *config.Config) *api.Client {
	c := api.NewClient(cfg)
	c.UserAgent = userAgent
	return c
}

// loadRuntimeConfig resolves the active auth source across flags/env/config.
// Tests override this to inject runtime-auth behavior without touching the real config file.
var loadRuntimeConfig = config.LoadRuntime

// loadConfig is the function used by command handlers to load configuration.
// Tests override this to inject test configs without touching the real config file.
var loadConfig = func() (*config.Config, error) {
	rt, err := loadRuntimeConfig()
	if err != nil {
		return nil, err
	}
	if rt == nil || rt.Effective == nil {
		return &config.Config{}, nil
	}
	return rt.Effective, nil
}
