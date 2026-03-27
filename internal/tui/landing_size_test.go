package tui

import (
	"net/http"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestLandingEnterPropagatesCurrentWindowSizeToFirstChild(t *testing.T) {
	client := newLandingTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})

	t.Run("markets", func(t *testing.T) {
		m := NewLandingModel(client)
		m.width = 132
		m.height = 44

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		lm := updated.(LandingModel)

		require.Equal(t, landingMarkets, lm.state)
		require.Equal(t, 132, lm.markets.width)
		require.Equal(t, 44, lm.markets.height)
		require.NotNil(t, cmd)
	})

	t.Run("trending", func(t *testing.T) {
		m := NewLandingModel(client)
		m.width = 132
		m.height = 44
		m.selection = 1

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		lm := updated.(LandingModel)

		require.Equal(t, landingTrending, lm.state)
		require.Equal(t, 50, lm.trending.limit)
		require.Equal(t, 132, lm.trending.width)
		require.Equal(t, 44, lm.trending.height)
		require.NotNil(t, cmd)
	})
}
