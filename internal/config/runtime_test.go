package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func withRuntimeIsolatedHome(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	t.Setenv("HOME", base)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(base, "xdg-config"))
	t.Setenv("APPDATA", filepath.Join(base, "AppData", "Roaming"))
	t.Setenv("USERPROFILE", base)
	return base
}

func TestLoadRuntime_EnvOnlyPreferredOverMissingConfig(t *testing.T) {
	withRuntimeIsolatedHome(t)
	t.Setenv("CMC_API_KEY", "env-runtime-key-123456")

	rt, err := LoadRuntime()
	require.NoError(t, err)
	require.NotNil(t, rt)
	require.Equal(t, AuthSourceEnv, rt.AuthSource)
	require.False(t, rt.Configured)
	require.NotNil(t, rt.Effective)
	require.Equal(t, "env-runtime-key-123456", rt.Effective.APIKey)
	require.Empty(t, rt.Effective.Tier)
	require.Empty(t, rt.ConfigWarning)
}

func TestLoadRuntime_ConfigBackedWhenEnvMissing(t *testing.T) {
	withRuntimeIsolatedHome(t)

	cfg := &Config{APIKey: "config-key-abcdefgh", Tier: TierStandard}
	require.NoError(t, Save(cfg))

	rt, err := LoadRuntime()
	require.NoError(t, err)
	require.NotNil(t, rt)
	require.Equal(t, AuthSourceConfig, rt.AuthSource)
	require.True(t, rt.Configured)
	require.NotNil(t, rt.Effective)
	require.Equal(t, cfg.APIKey, rt.Effective.APIKey)
	require.Equal(t, cfg.Tier, rt.Effective.Tier)
	require.Empty(t, rt.ConfigWarning)
}

func TestLoadRuntime_InvalidPersistedTierProducesWarning(t *testing.T) {
	withRuntimeIsolatedHome(t)

	path, err := ConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte("api_key: invalid-tier-key\ntier: nope\n"), 0o600))

	rt, err := LoadRuntime()
	require.NoError(t, err)
	require.NotNil(t, rt)
	require.Equal(t, AuthSourceConfig, rt.AuthSource)
	require.True(t, rt.Configured)
	require.NotNil(t, rt.Effective)
	require.Equal(t, "invalid-tier-key", rt.Effective.APIKey)
	require.Empty(t, rt.Effective.Tier)
	require.NotEmpty(t, rt.Warning)
	require.NotEmpty(t, rt.ConfigWarning)
	require.True(t, strings.Contains(strings.ToLower(rt.Warning), "tier"))
}
