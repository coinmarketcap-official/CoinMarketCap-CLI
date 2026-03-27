package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDEX_NotListedInRootHelp(t *testing.T) {
	stdout, stderr, err := executeCommandCLI(t, "--help")
	require.NoError(t, err)

	combined := stdout + stderr
	assert.False(t, helpListsCommandUnderAvailableCommands(combined, "dex"),
		"public root help must not list dex after the surface cleanup")
}

func TestDEX_NotEmittedInCommandsCatalog(t *testing.T) {
	stdout, _, err := executeCommand(t, "commands", "-o", "json")
	require.NoError(t, err)

	var catalog commandCatalog
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout)), &catalog))

	for _, cmd := range catalog.Commands {
		assert.False(t, cmd.Name == "dex" || strings.HasPrefix(cmd.Name, "dex "),
			"commands catalog must not expose public dex entries, found %q", cmd.Name)
	}
}

func TestDEX_PublicInvocationFailsAsUnknownCommand(t *testing.T) {
	stdout, stderr, err := executeCommand(t, "dex")
	require.Error(t, err)
	assert.Empty(t, strings.TrimSpace(stdout))

	combined := stderr + err.Error()
	assert.Contains(t, combined, "unknown command")
	assert.Contains(t, combined, "dex")
}
