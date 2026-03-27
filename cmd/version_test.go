package cmd

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Keys contract for `cmc version -o json`.
var versionJSONStableKeys = []string{"version", "commit", "date", "go", "os", "arch"}

func TestVersion_JSON_HasStableKeys(t *testing.T) {
	stdout, _, err := executeCommand(t, "version", "-o", "json")
	require.NoError(t, err)

	var m map[string]string
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &m))

	for _, k := range versionJSONStableKeys {
		require.Contains(t, m, k, "expected stable key %q in JSON output", k)
	}

	require.Equal(t, version, m["version"], "JSON version should match package version var")
	require.Equal(t, commit, m["commit"], "JSON commit should match package commit var")
	require.Equal(t, date, m["date"], "JSON date should match package date var")

	require.NotEmpty(t, m["go"], "JSON go should be non-empty")
	require.NotEmpty(t, m["os"], "JSON os should be non-empty")
	require.NotEmpty(t, m["arch"], "JSON arch should be non-empty")
	require.Equal(t, runtime.Version(), m["go"])
	require.Equal(t, runtime.GOOS, m["os"])
	require.Equal(t, runtime.GOARCH, m["arch"])
}

func TestVersion_Table_CMCBranding(t *testing.T) {
	stdout, _, err := executeCommand(t, "version", "-o", "table")
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	require.NotEmpty(t, lines)

	first := lines[0]
	require.True(t, strings.HasPrefix(first, "cmc "), "first line should use cmc branding, got %q", first)
	require.False(t, strings.HasPrefix(first, "cg "), "first line must not use legacy cg branding")

	out := strings.TrimSpace(stdout)
	require.Contains(t, out, "commit:", "table output should include commit label")
	require.Contains(t, out, "built:", "table output should include built label")
	require.Contains(t, out, "go:", "table output should include go label")
	require.Contains(t, out, "os:", "table output should include os label")
}
