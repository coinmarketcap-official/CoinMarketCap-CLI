package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const RefreshInterval = 60 * time.Second

type refreshTickMsg struct{}

func refreshTick() tea.Cmd {
	return tea.Tick(RefreshInterval, func(time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}
