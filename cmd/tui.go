package cmd

import (
	"fmt"

	"github.com/openCMC/CoinMarketCap-CLI/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type tuiProgram interface {
	Run() (tea.Model, error)
}

var newTUIProgram = func(model tea.Model, opts ...tea.ProgramOption) tuiProgram {
	return tea.NewProgram(model, opts...)
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI for browsing crypto data",
	Example: `  cmc tui
  cmc tui markets
  cmc tui trending`,
	RunE: runTUILanding,
}

var tuiMarketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "Browse top coins by market cap",
	Example: `  cmc tui markets
  cmc tui markets --category layer-2
  cmc tui markets --convert EUR`,
	RunE: runTUIMarkets,
}

var tuiTrendingCmd = &cobra.Command{
	Use:   "trending",
	Short: "Browse trending coins",
	Example: `  cmc tui trending
  cmc tui trending --convert EUR`,
	RunE: runTUITrending,
}

func init() {
	tuiMarketsCmd.Flags().Int("total", 50, "Total number of coins to fetch")
	tuiMarketsCmd.Flags().String("category", "", "Filter by CoinMarketCap category slug")
	tuiMarketsCmd.Flags().String("convert", "USD", "Target currency")
	tuiTrendingCmd.Flags().String("convert", "USD", "Target currency")

	tuiCmd.AddCommand(tuiMarketsCmd)
	tuiCmd.AddCommand(tuiTrendingCmd)
	rootCmd.AddCommand(tuiCmd)
}

func runTUILanding(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client := newAPIClient(cfg)
	model := tui.NewLandingModel(client)

	p := newTUIProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func runTUIMarkets(cmd *cobra.Command, args []string) error {
	total, _ := cmd.Flags().GetInt("total")
	category, _ := cmd.Flags().GetString("category")
	convert, _ := cmd.Flags().GetString("convert")

	if total <= 0 {
		return fmt.Errorf("--total must be a positive integer")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client := newAPIClient(cfg)
	model := tui.NewMarketsModelWithCategory(client, convert, total, category)

	p := newTUIProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func runTUITrending(cmd *cobra.Command, args []string) error {
	convert, _ := cmd.Flags().GetString("convert")

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client := newAPIClient(cfg)
	model := tui.NewTrendingModel(client, convert)

	p := newTUIProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
