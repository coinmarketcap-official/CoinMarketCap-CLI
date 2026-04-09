package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"
	"github.com/openCMC/CoinMarketCap-CLI/internal/config"
	"github.com/stretchr/testify/require"
)

// withIsolatedUserConfigDir redirects os.UserConfigDir resolution so config.Load/Save
// never touch the real user profile. Platform notes:
//   - darwin: HOME → temp (UserConfigDir is $HOME/Library/Application Support)
//   - Unix (non-darwin): HOME + XDG_CONFIG_HOME (inherits XDG from CI would bypass HOME alone)
//   - windows: APPDATA → temp Roaming-style path (USERPROFILE alone does not clear APPDATA)
func withIsolatedUserConfigDir(t *testing.T) {
	t.Helper()
	base := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		roaming := filepath.Join(base, "AppData", "Roaming")
		t.Setenv("APPDATA", roaming)
		t.Setenv("USERPROFILE", base)
	case "darwin":
		t.Setenv("HOME", base)
	default:
		t.Setenv("HOME", base)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "xdg-config"))
	}
}

func TestAuth_NonInteractiveEnvSavesConfig(t *testing.T) {
	withIsolatedUserConfigDir(t)

	const wantKey = "env-secret-api-key-12345"
	t.Setenv("CMC_API_KEY", wantKey)

	_, _, err := executeCommandCLI(t, "auth", "--skip-verify")
	if err != nil {
		t.Fatalf("auth: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIKey != wantKey {
		t.Errorf("APIKey: got %q, want %q", cfg.APIKey, wantKey)
	}
	if cfg.Tier != "" {
		t.Errorf("Tier: got %q, want empty tier", cfg.Tier)
	}

	path, err := config.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file should exist at %s: %v", path, err)
	}
}

func TestAuth_InvalidEnvTierIsIgnored(t *testing.T) {
	withIsolatedUserConfigDir(t)

	const wantKey = "test-key"
	t.Setenv("CMC_API_KEY", "test-key")
	t.Setenv("CMC_API_TIER", "not-a-real-tier")

	_, _, err := executeCommandCLI(t, "auth", "--skip-verify")
	require.NoError(t, err)

	cfg, loadErr := config.Load()
	require.NoError(t, loadErr)
	require.Equal(t, wantKey, cfg.APIKey)
	require.Empty(t, cfg.Tier)
}

func TestAuth_KeyFlagShellHistoryWarning(t *testing.T) {
	withIsolatedUserConfigDir(t)

	// No CMC_API_KEY — key supplied only via flag so stderr warning path runs.

	const flagKey = "flag-supplied-secret-key"
	_, stderr, err := executeCommandCLI(t, "auth", "--key", flagKey, "--skip-verify")
	if err != nil {
		t.Fatalf("auth: %v", err)
	}

	if !strings.Contains(stderr, "shell history") {
		t.Errorf("stderr should warn about shell history; got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "--key") {
		t.Errorf("stderr should mention --key flag; got:\n%s", stderr)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIKey != flagKey {
		t.Errorf("APIKey: got %q, want %q", cfg.APIKey, flagKey)
	}
	if cfg.Tier != "" {
		t.Errorf("Tier: got %q, want empty tier", cfg.Tier)
	}
}

func TestAuth_FlagKeyWinsOverEnvKey(t *testing.T) {
	withIsolatedUserConfigDir(t)

	const envKey = "env-wins-api-key"
	const flagKey = "flag-should-not-persist"
	t.Setenv("CMC_API_KEY", envKey)

	_, _, err := executeCommandCLI(t, "auth", "--key", flagKey, "--skip-verify")
	if err != nil {
		t.Fatalf("auth: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIKey != flagKey {
		t.Errorf("APIKey: got %q, want %q (flags must override env)", cfg.APIKey, flagKey)
	}
	if cfg.Tier != "" {
		t.Errorf("Tier: got %q, want empty tier", cfg.Tier)
	}
}

func TestAuth_InvalidKeyFailsVerification(t *testing.T) {
	withIsolatedUserConfigDir(t)

	const badKey = "definitely-invalid-key"
	t.Setenv("CMC_API_KEY", badKey)

	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-CMC_PRO_API_KEY"); got != badKey {
			t.Fatalf("verification request missing API key header: got %q", got)
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"status":{"error_code":1001,"error_message":"Invalid API Key."}}`)
	})
	defer srv.Close()

	origClient := newAPIClient
	newAPIClient = func(cfg *config.Config) *api.Client {
		c := api.NewClient(cfg)
		c.SetBaseURL(srv.URL)
		return c
	}
	t.Cleanup(func() { newAPIClient = origClient })

	_, stderr, err := executeCommandCLI(t, "auth")
	if err == nil {
		t.Fatal("expected auth to fail for obviously invalid key")
	}
	if !strings.Contains(strings.ToLower(stderr), "invalid") {
		t.Fatalf("stderr should explain invalid auth; got:\n%s", stderr)
	}

	path, errPath := config.ConfigPath()
	if errPath != nil {
		t.Fatalf("ConfigPath: %v", errPath)
	}
	if _, statErr := os.Stat(path); statErr == nil {
		t.Fatalf("config file should not be written when verification fails, but %s exists", path)
	}
}

func TestAuth_TransientVerificationStillSavesUnverified(t *testing.T) {
	withIsolatedUserConfigDir(t)

	const wantKey = "transient-save-key-12345"
	t.Setenv("CMC_API_KEY", wantKey)

	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-CMC_PRO_API_KEY"); got != wantKey {
			t.Fatalf("verification request missing API key header: got %q", got)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"status":{"error_code":500,"error_message":"Server error."}}`)
	})
	defer srv.Close()

	origClient := newAPIClient
	newAPIClient = func(cfg *config.Config) *api.Client {
		c := api.NewClient(cfg)
		c.SetBaseURL(srv.URL)
		return c
	}
	t.Cleanup(func() { newAPIClient = origClient })

	_, stderr, err := executeCommandCLI(t, "auth")
	if err != nil {
		t.Fatalf("auth should save even when verification is unavailable: %v", err)
	}
	if !strings.Contains(strings.ToLower(stderr), "unverified") {
		t.Fatalf("stderr should state that the save was unverified; got:\n%s", stderr)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIKey != wantKey {
		t.Fatalf("saved key mismatch: got %q want %q", cfg.APIKey, wantKey)
	}
	if cfg.Tier != "" {
		t.Fatalf("saved tier mismatch: got %q want empty tier", cfg.Tier)
	}
}
