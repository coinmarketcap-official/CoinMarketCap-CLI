package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoot_Version_PrintsVersionLine(t *testing.T) {
	stdout, stderr, err := executeCommandCLI(t, "--version")
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(stderr), "root --version should not write diagnostics to stderr")
	line := strings.TrimSpace(stdout)
	assert.Contains(t, line, "cmc")
	assert.Contains(t, line, rootCmd.Version)
	assert.Contains(t, line, "version")
}

func TestRoot_NoArgs_PrintsWelcomeSurface(t *testing.T) {
	stdout, stderr, err := executeCommand(t)
	require.NoError(t, err)

	out := stderr
	assert.Contains(t, out, "CoinMarketCap CLI", "logo line should appear on stderr")
	assert.Contains(t, out, "Quick Start", "welcome box should list quick start section")
	assert.Contains(t, out, "cmc auth", "welcome box should show sample commands")
	assert.Empty(t, strings.TrimSpace(stdout), "welcome surface should not write to stdout")
}

func TestHelp_Root_RendersHelpText(t *testing.T) {
	stdout, stderr, err := executeCommand(t, "--help")
	require.NoError(t, err)

	long := strings.TrimSpace(rootCmd.Long)
	require.NotEmpty(t, long, "rootCmd.Long must be non-empty for help blurb checks to be meaningful")
	// Cobra writes help to stdout when using default OutWriter.
	assert.Contains(t, stdout, long)
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "Available Commands:")
	assert.Empty(t, strings.TrimSpace(stderr), "root --help should not write diagnostics to stderr")
}

func TestHelp_Root_DoesNotListHiddenCommandsSubcommand(t *testing.T) {
	stdout, stderr, err := executeCommand(t, "--help")
	require.NoError(t, err)

	combined := stdout + stderr
	require.Contains(t, combined, "Available Commands:", "precondition: help should include public command list")
	assert.False(t, helpListsCommandUnderAvailableCommands(combined, "commands"),
		"hidden 'commands' subcommand must not appear in Available Commands list")
}

// helpListsCommandUnderAvailableCommands reports whether help lists name as a
// top-level subcommand under "Available Commands:" (first word of a non-empty
// line in that section). Tolerates varying indentation.
func helpListsCommandUnderAvailableCommands(help, name string) bool {
	const hdr = "Available Commands:"
	i := strings.Index(help, hdr)
	if i < 0 {
		return false
	}
	body := help[i+len(hdr):]
	end := len(body)
	for _, sep := range []string{"\nFlags:", "\nGlobal Flags:"} {
		if j := strings.Index(body, sep); j >= 0 && j < end {
			end = j
		}
	}
	body = body[:end]
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return true
		}
	}
	return false
}
