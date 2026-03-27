package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withStubLoadConfig(t *testing.T, cfg *config.Config) {
	t.Helper()

	orig := loadConfig
	loadConfig = func() (*config.Config, error) {
		return cfg, nil
	}
	t.Cleanup(func() { loadConfig = orig })
}

func withCapturedTUIProgram(t *testing.T) {
	t.Helper()

	orig := newTUIProgram
	newTUIProgram = func(model tea.Model, opts ...tea.ProgramOption) tuiProgram {
		return captureProgram{model: model}
	}
	t.Cleanup(func() { newTUIProgram = orig })
}

func withAuthContractEnv(t *testing.T) {
	t.Helper()

	withIsolatedUserConfigDir(t)
	t.Setenv("CMC_API_KEY", "auth-contract-key")
	t.Setenv("CMC_API_TIER", config.TierStandard)
}

func commandInfoByName(t *testing.T, cat commandCatalog, name string) commandInfo {
	t.Helper()

	for i := range cat.Commands {
		if cat.Commands[i].Name == name {
			return cat.Commands[i]
		}
	}
	t.Fatalf("catalog missing command %q", name)
	return commandInfo{}
}

func flagNameSet(info commandInfo) map[string]struct{} {
	out := make(map[string]struct{}, len(info.Flags))
	for _, f := range info.Flags {
		out[f.Name] = struct{}{}
	}
	return out
}

func requireFlagPresence(t *testing.T, info commandInfo, wantPresent []string, wantAbsent []string) {
	t.Helper()

	flags := flagNameSet(info)
	for _, name := range wantPresent {
		_, ok := flags[name]
		require.True(t, ok, "%q must advertise flag %q", info.Name, name)
	}
	for _, name := range wantAbsent {
		_, ok := flags[name]
		require.False(t, ok, "%q must not advertise flag %q", info.Name, name)
	}
}

func TestPublicHelp_SurfaceFlagContracts(t *testing.T) {
	cases := []struct {
		name           string
		args           []string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "auth",
			args:           []string{"auth", "--help"},
			wantNotContain: []string{"--dry-run", "-o, --output"},
		},
		{
			name:           "status",
			args:           []string{"status", "--help"},
			wantContains:   []string{"-o, --output"},
			wantNotContain: []string{"--dry-run"},
		},
		{
			name:           "version",
			args:           []string{"version", "--help"},
			wantContains:   []string{"-o, --output"},
			wantNotContain: []string{"--dry-run"},
		},
		{
			name:           "tui",
			args:           []string{"tui", "--help"},
			wantNotContain: []string{"--dry-run", "-o, --output"},
		},
		{
			name:           "tui markets",
			args:           []string{"tui", "markets", "--help"},
			wantNotContain: []string{"--dry-run", "-o, --output"},
		},
		{
			name:           "tui trending",
			args:           []string{"tui", "trending", "--help"},
			wantNotContain: []string{"--dry-run", "-o, --output"},
		},
		{
			name:         "metrics",
			args:         []string{"metrics", "--help"},
			wantContains:  []string{"--convert", "-o, --output", "--dry-run"},
		},
		{
			name:         "news",
			args:         []string{"news", "--help"},
			wantContains:  []string{"--start", "--limit", "--language", "--news-type", "-o, --output", "--dry-run"},
		},
		{
			name:         "pairs",
			args:         []string{"pairs", "--help"},
			wantContains:  []string{"--limit", "--category", "--convert", "-o, --output", "--dry-run"},
		},
		{
			name:         "price",
			args:         []string{"price", "--help"},
			wantContains:  []string{"--with-info", "--with-chain-stats", "--convert", "-o, --output", "--dry-run"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := executeCommandCLI(t, tc.args...)
			require.NoError(t, err)

			combined := stdout + stderr
			for _, want := range tc.wantContains {
				assert.Contains(t, combined, want)
			}
			for _, want := range tc.wantNotContain {
				assert.NotContains(t, combined, want)
			}
		})
	}
}

func TestPublicFlagUse_UnsupportedFlagsError(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		setup func(*testing.T)
	}{
		{
			name:  "auth dry-run",
			args:  []string{"auth", "--dry-run"},
			setup: withAuthContractEnv,
		},
		{
			name:  "auth output",
			args:  []string{"auth", "-o", "json"},
			setup: withAuthContractEnv,
		},
		{
			name: "status dry-run",
			args: []string{"status", "--dry-run"},
			setup: func(t *testing.T) {
				withStubLoadConfig(t, &config.Config{Tier: config.TierStandard})
			},
		},
		{
			name: "version dry-run",
			args: []string{"version", "--dry-run"},
		},
		{
			name: "tui dry-run",
			args: []string{"tui", "--dry-run"},
			setup: func(t *testing.T) {
				withStubLoadConfig(t, &config.Config{Tier: config.TierEnterprise})
				withCapturedTUIProgram(t)
			},
		},
		{
			name: "tui output",
			args: []string{"tui", "-o", "json"},
			setup: func(t *testing.T) {
				withStubLoadConfig(t, &config.Config{Tier: config.TierEnterprise})
				withCapturedTUIProgram(t)
			},
		},
		{
			name: "tui markets dry-run",
			args: []string{"tui", "markets", "--dry-run"},
			setup: func(t *testing.T) {
				withStubLoadConfig(t, &config.Config{Tier: config.TierEnterprise})
				withCapturedTUIProgram(t)
			},
		},
		{
			name: "tui markets output",
			args: []string{"tui", "markets", "-o", "json"},
			setup: func(t *testing.T) {
				withStubLoadConfig(t, &config.Config{Tier: config.TierEnterprise})
				withCapturedTUIProgram(t)
			},
		},
		{
			name: "tui trending dry-run",
			args: []string{"tui", "trending", "--dry-run"},
			setup: func(t *testing.T) {
				withStubLoadConfig(t, &config.Config{Tier: config.TierEnterprise})
				withCapturedTUIProgram(t)
			},
		},
		{
			name: "tui trending output",
			args: []string{"tui", "trending", "-o", "json"},
			setup: func(t *testing.T) {
				withStubLoadConfig(t, &config.Config{Tier: config.TierEnterprise})
				withCapturedTUIProgram(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t)
			}

			_, _, err := executeCommandCLI(t, tc.args...)
			require.Error(t, err)
		})
	}
}

func TestHelp_ResolveAndNewsExamplesAreCopyPasteSafe(t *testing.T) {
	resolveHelp, _, err := executeCommand(t, "resolve", "--help")
	require.NoError(t, err)
	assert.NotContains(t, resolveHelp, "cmc resolve --symbol BTC")
	assert.Contains(t, resolveHelp, "cmc resolve --id 1")
	assert.Contains(t, resolveHelp, "cmc resolve --slug bitcoin")

	newsHelp, _, err := executeCommand(t, "news", "--help")
	require.NoError(t, err)
	assert.NotContains(t, newsHelp, "news-type top")
	assert.Contains(t, newsHelp, "cmc news --start 21 --language en --news-type news")
}

func TestReadme_DoesNotContainBrokenExamples(t *testing.T) {
	readmePath := "../README.md"
	b, err := os.ReadFile(readmePath)
	require.NoError(t, err)

	readme := string(b)
	assert.NotContains(t, readme, "cmc resolve --symbol BTC")
	assert.NotContains(t, readme, "cmc news --language en --news-type top -o table")
	assert.Contains(t, readme, "cmc resolve --id 1")
	assert.Contains(t, readme, "cmc news --language en --news-type news -o table")
}

func TestCommandsCatalog_PublicSurfaceFlagsAndFormats(t *testing.T) {
	cat := parseCommandsJSONCatalog(t)

	auth := commandInfoByName(t, cat, "auth")
	require.Empty(t, auth.OutputFormats, "auth should not advertise output formats")
	requireFlagPresence(t, auth, nil, []string{"output", "dry-run"})

	status := commandInfoByName(t, cat, "status")
	require.Equal(t, []string{"json", "table"}, status.OutputFormats)
	requireFlagPresence(t, status, []string{"output"}, []string{"dry-run"})

	version := commandInfoByName(t, cat, "version")
	require.Equal(t, []string{"table", "json"}, version.OutputFormats)
	requireFlagPresence(t, version, []string{"output"}, []string{"dry-run"})

	for _, name := range []string{"markets", "history", "resolve", "search", "trending", "top-gainers-losers", "monitor", "metrics", "news", "pairs"} {
		entry := commandInfoByName(t, cat, name)
		requireFlagPresence(t, entry, []string{"output", "dry-run"}, nil)
	}

	price := commandInfoByName(t, cat, "price")
	requireFlagPresence(t, price, []string{"output", "dry-run", "with-chain-stats"}, nil)

	for _, name := range []string{"tui markets", "tui trending"} {
		entry := commandInfoByName(t, cat, name)
		requireFlagPresence(t, entry, nil, []string{"output", "dry-run"})
		require.Equal(t, []string{"tui"}, entry.OutputFormats)
	}
}

func TestHistory_DryRun_BasicTierStillReturnsPreview(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("dry-run should not make HTTP requests")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommandCLI(t, "history", "--id", "1", "--days", "7", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &out))
	assert.Contains(t, out.URL, "/v1/cryptocurrency/quotes/historical")
	assert.Equal(t, "1", out.Params["id"])
	assert.Equal(t, "USD", out.Params["convert"])
	assert.Equal(t, "daily", out.Params["interval"])
}

func TestPriceCatalog_ExposesWithInfoEndpoint(t *testing.T) {
	cat := parseCommandsJSONCatalog(t)
	price := commandInfoByName(t, cat, "price")
	require.Contains(t, price.APIEndpoints, "with-info")
	require.Equal(t, "/v2/cryptocurrency/info", price.APIEndpoints["with-info"])
	require.Contains(t, price.APIEndpoints, "with-chain-stats")
	require.Equal(t, "/v1/blockchain/statistics/latest", price.APIEndpoints["with-chain-stats"])
	require.Contains(t, price.OASOperationIDs, "with-info")
	require.Equal(t, "getV2CryptocurrencyInfo", price.OASOperationIDs["with-info"])
	require.Contains(t, price.OASOperationIDs, "with-chain-stats")
	require.Equal(t, "getV1BlockchainStatisticsLatest", price.OASOperationIDs["with-chain-stats"])
}

func TestPublicSurface_NewCommandsAppearInHelpAndCatalog(t *testing.T) {
	cat := parseCommandsJSONCatalog(t)

	for _, name := range []string{"metrics", "news", "pairs"} {
		entry := commandInfoByName(t, cat, name)
		requireFlagPresence(t, entry, []string{"output", "dry-run"}, nil)
		require.NotEmpty(t, entry.APIEndpoint, "%q must surface an API endpoint in the catalog", name)
		require.NotEmpty(t, entry.OASOperationID, "%q must surface an operation id in the catalog", name)
	}
}
