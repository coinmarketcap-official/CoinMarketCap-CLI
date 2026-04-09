package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"
	"github.com/openCMC/CoinMarketCap-CLI/internal/config"
	"github.com/openCMC/CoinMarketCap-CLI/internal/display"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure API key",
	Example: `  cmc auth
  CMC_API_KEY=your-key cmc auth`,
	RunE: runAuth,
}

func init() {
	authCmd.Flags().String("key", "", "CoinMarketCap API key")
	authCmd.Flags().Bool("skip-verify", false, "Save credentials without verifying them first")
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) error {
	display.PrintBanner()
	ctx := cmd.Context()

	key, _ := cmd.Flags().GetString("key")
	skipVerify, _ := cmd.Flags().GetBool("skip-verify")

	if cmd.Flags().Changed("key") {
		warnf("Warning: --key flag exposes secrets in shell history. Prefer CMC_API_KEY env var or interactive prompt.\n")
	}

	// Prefer flags over env to preserve explicit CLI input.
	if !cmd.Flags().Changed("key") {
		if envKey := os.Getenv("CMC_API_KEY"); envKey != "" {
			key = envKey
		}
	}
	if key == "" {
		if err := huh.NewInput().
			Title("API Key").
			Description("Enter your CoinMarketCap API key").
			EchoMode(huh.EchoModePassword).
			Value(&key).
			Run(); err != nil {
			return err
		}
	}

	cfg := &config.Config{APIKey: strings.TrimSpace(key)}

	if skipVerify {
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		warnf("Saved locally (verification skipped). Key: %s\n", cfg.MaskedKey())
		return nil
	}

	if verifyErr := verifyAuth(ctx, cfg); verifyErr != nil {
		if errors.Is(verifyErr, api.ErrInvalidAPIKey) {
			warnf("Authentication failed: %v\n", verifyErr)
			return verifyErr
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		warnf("Saved locally (verification unverified): %v\n", verifyErr)
		return nil
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	warnf("Saved and verified. Key: %s\n", cfg.MaskedKey())
	return nil
}

func verifyAuth(ctx context.Context, cfg *config.Config) error {
	client := newAPIClient(cfg)
	_, err := client.InfoByID(ctx, "1")
	return err
}
