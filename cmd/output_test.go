package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyError_RateLimited(t *testing.T) {
	rle := &api.RateLimitError{RetryAfter: 30}
	ce := classifyError(rle)
	assert.Equal(t, "rate_limited", ce.Error)
	assert.NotNil(t, ce.RetryAfter)
	assert.Equal(t, 30, *ce.RetryAfter)
}

func TestClassifyError_RateLimitedNoRetryAfter(t *testing.T) {
	rle := &api.RateLimitError{}
	ce := classifyError(rle)
	assert.Equal(t, "rate_limited", ce.Error)
	assert.Nil(t, ce.RetryAfter)
}

func TestClassifyError_InvalidAPIKey(t *testing.T) {
	ce := classifyError(api.ErrInvalidAPIKey)
	assert.Equal(t, "invalid_api_key", ce.Error)
}

func TestClassifyError_PlanRestricted(t *testing.T) {
	ce := classifyError(api.ErrPlanRestricted)
	assert.Equal(t, "plan_restricted", ce.Error)
}

func TestClassifyError_GenericError(t *testing.T) {
	ce := classifyError(assert.AnError)
	assert.Equal(t, "error", ce.Error)
	assert.Nil(t, ce.RetryAfter)
}

func TestFormatError_JSONMode_WritesStderr(t *testing.T) {
	// Set up a command with -o json flag, then call formatError directly.
	// This verifies JSON is written to stderr in JSON output mode.
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Prepare a command with -o json set.
	resetAllFlags(rootCmd)
	rootCmd.SetArgs([]string{"resolve", "-o", "json"})
	cmd, _, _ := rootCmd.Find([]string{"resolve", "-o", "json"})
	require.NotNil(t, cmd)
	_ = cmd.Flags().Set("output", "json")

	err := formatError(cmd, api.ErrInvalidAPIKey)
	require.Error(t, err) // formatError returns the original error

	_ = wErr.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(rErr)
	os.Stderr = oldStderr

	var cliErr CLIError
	require.NoError(t, json.Unmarshal(buf.Bytes(), &cliErr), "stderr should contain valid JSON: %s", buf.String())
	assert.Equal(t, "invalid_api_key", cliErr.Error)
	assert.NotContains(t, buf.String(), "\n  ")
}

func TestFormatError_TableMode_NoJSON(t *testing.T) {
	// In table mode, formatError should return the error without writing JSON.
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	resetAllFlags(rootCmd)
	rootCmd.SetArgs([]string{"resolve"})
	cmd, _, _ := rootCmd.Find([]string{"resolve"})
	require.NotNil(t, cmd)
	_ = cmd.Flags().Set("output", "table")

	err := formatError(cmd, api.ErrInvalidAPIKey)
	require.Error(t, err)

	_ = wErr.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(rErr)
	os.Stderr = oldStderr

	// stderr should be empty — formatError only writes JSON in JSON mode.
	assert.Empty(t, buf.String(), "formatError should not write to stderr in table mode")
}

func TestExecuteCommand_InvalidOutputFormat_FailsLocally(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call when output validation fails")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommandCLI(t, "price", "btc", "-o", "xml")
	require.Error(t, err)
	assert.Empty(t, stdout)
	assert.Contains(t, err.Error(), "--output must be one of: json, table")
}

func TestClassifyError_AmbiguousSymbol_IncludesCandidates(t *testing.T) {
	// When resolving an ambiguous symbol, the JSON error must include candidates.
	ambigErr := &api.ResolverAmbiguityError{
		Symbol: "BTC",
		Candidates: []api.ResolvedAsset{
			{ID: 1, Slug: "bitcoin", Symbol: "BTC", Name: "Bitcoin", Rank: 1, IsActive: true},
			{ID: 2, Slug: "bitcoin-cash", Symbol: "BTC", Name: "Bitcoin Cash", Rank: 5, IsActive: true},
		},
	}

	ce := classifyError(ambigErr)
	assert.Equal(t, "resolver_ambiguous", ce.Error)
	assert.Contains(t, ce.Message, "BTC")

	// Candidates must be included in the structured error.
	require.NotNil(t, ce.Candidates, "candidates must be present in JSON error")
	assert.Len(t, ce.Candidates, 2)
	assert.Equal(t, int64(1), ce.Candidates[0].ID)
	assert.Equal(t, "bitcoin", ce.Candidates[0].Slug)
}
