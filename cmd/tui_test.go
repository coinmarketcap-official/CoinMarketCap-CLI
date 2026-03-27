package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/config"
	"github.com/coinmarketcap/coinmarketcap-cli/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

type captureProgram struct {
	model tea.Model
}

func (p captureProgram) Run() (tea.Model, error) {
	return p.model, nil
}

func TestTUICommandEntryPoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()

	withTestClient(t, srv, config.TierEnterprise)

	var captured tea.Model
	orig := newTUIProgram
	newTUIProgram = func(model tea.Model, opts ...tea.ProgramOption) tuiProgram {
		captured = model
		return captureProgram{model: model}
	}
	t.Cleanup(func() { newTUIProgram = orig })

	_, _, err := executeCommandCLI(t, "tui")
	require.NoError(t, err)
	_, ok := captured.(tui.LandingModel)
	require.True(t, ok, "cmc tui should open the landing model")

	captured = nil
	_, _, err = executeCommandCLI(t, "tui", "markets")
	require.NoError(t, err)
	_, ok = captured.(tui.MarketsModel)
	require.True(t, ok, "cmc tui markets should open the markets model")

	captured = nil
	_, _, err = executeCommandCLI(t, "tui", "trending")
	require.NoError(t, err)
	_, ok = captured.(tui.TrendingModel)
	require.True(t, ok, "cmc tui trending should open the trending model")
}
