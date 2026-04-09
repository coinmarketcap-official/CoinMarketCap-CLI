package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/openCMC/CoinMarketCap-CLI/internal/api"
	"github.com/openCMC/CoinMarketCap-CLI/internal/config"
	"github.com/openCMC/CoinMarketCap-CLI/internal/display"

	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve assets by id, slug, or symbol",
	Example: `  cmc resolve --id 1
  cmc resolve --slug bitcoin`,
	Args: cobra.NoArgs,
	RunE: runResolve,
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search assets or discover chain-scoped pair/token addresses",
	Example: `  cmc search bitcoin
  cmc search btc --limit 5
  cmc search --chain ethereum --address 0xabc123`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSearch,
}

func init() {
	resolveCmd.Flags().String("id", "", "Resolve an exact CoinMarketCap numeric ID")
	resolveCmd.Flags().String("slug", "", "Resolve an exact slug")
	resolveCmd.Flags().String("symbol", "", "Resolve a symbol, failing on ambiguity")
	addOutputFlag(resolveCmd)
	addDryRunFlag(resolveCmd)
	rootCmd.AddCommand(resolveCmd)

	searchCmd.Flags().Int("limit", 10, "Maximum number of results to return")
	searchCmd.Flags().String("chain", "", "Chain name or numeric id (required with --address)")
	searchCmd.Flags().String("address", "", "Contract or pair address to discover within a specific chain (pair lookup first, then paged spot-pairs scan)")
	addOutputFlag(searchCmd)
	addDryRunFlag(searchCmd)
	rootCmd.AddCommand(searchCmd)
}

func runResolve(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetString("id")
	slug, _ := cmd.Flags().GetString("slug")
	symbol, _ := cmd.Flags().GetString("symbol")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	provided := 0
	for _, v := range []string{id, slug, symbol} {
		if v != "" {
			provided++
		}
	}
	if provided != 1 {
		return fmt.Errorf("specify exactly one of --id, --slug, or --symbol")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if isDryRun(cmd) {
		params := map[string]string{}
		endpoint := "/v1/cryptocurrency/map"
		opKey := "id"
		if id != "" {
			params["id"] = id
			endpoint = "/v2/cryptocurrency/info"
			opKey = "id"
		}
		if slug != "" {
			params["slug"] = slug
			endpoint = "/v2/cryptocurrency/info"
			opKey = "slug"
		}
		if symbol != "" {
			params["symbol"] = symbol
			opKey = "symbol"
		}
		return printDryRunWithOp(cfg, "resolve", opKey, endpoint, params, nil)
	}

	client := newAPIClient(cfg)
	ctx := cmd.Context()
	var asset api.ResolvedAsset
	switch {
	case id != "":
		asset, err = client.ResolveByID(ctx, id)
	case slug != "":
		asset, err = client.ResolveBySlug(ctx, slug)
	default:
		asset, err = client.ResolveBySymbol(ctx, symbol)
	}
	if err != nil {
		var ambig *api.ResolverAmbiguityError
		if errors.As(err, &ambig) {
			if jsonOut {
				_ = json.NewEncoder(os.Stderr).Encode(classifyError(err))
			} else {
				renderResolveCandidates(ambig)
			}
			return markErrorRendered(err)
		}
		return err
	}
	assets := []api.ResolvedAsset{asset}

	if jsonOut {
		return printJSONRaw(assets)
	}

	headers := []string{"ID", "Slug", "Symbol", "Name", "Rank", "Active"}
	rows := make([][]string, len(assets))
	for i, asset := range assets {
		rows[i] = []string{
			fmt.Sprintf("%d", asset.ID),
			display.SanitizeCell(asset.Slug),
			display.SanitizeCell(asset.Symbol),
			display.SanitizeCell(asset.Name),
			formatResolvedRank(asset),
			formatResolvedActive(asset),
		}
	}

	display.PrintTable(headers, rows)
	return nil
}

func runSearch(cmd *cobra.Command, args []string) error {
	jsonOut := outputJSON(cmd)
	if !jsonOut {
		display.PrintBanner()
	}

	query, chain, address, limit, err := parseSearchInputs(cmd, args)
	if err != nil {
		return err
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if isDryRun(cmd) {
		if address != "" {
			return printAddressSearchDryRun(cfg, chain, address, limit)
		}
		params := map[string]string{
			"limit": strconv.Itoa(limit),
		}
		return printDryRun(cfg, "search", "/v1/cryptocurrency/map", params, nil)
	}

	client := newAPIClient(cfg)
	ctx := cmd.Context()

	var results []api.SearchResult
	switch {
	case address != "":
		results, err = client.SearchByAddress(ctx, chain, address, limit)
	default:
		results, err = client.SearchAssets(ctx, query, limit)
	}
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(results)
	}

	renderSearchTable(results)
	return nil
}

func parseSearchInputs(cmd *cobra.Command, args []string) (query, chain, address string, limit int, err error) {
	chain, _ = cmd.Flags().GetString("chain")
	address, _ = cmd.Flags().GetString("address")
	limit, _ = cmd.Flags().GetInt("limit")

	chain = strings.TrimSpace(chain)
	address = strings.TrimSpace(address)
	if limit <= 0 {
		return "", "", "", 0, fmt.Errorf("--limit must be greater than 0")
	}

	if len(args) > 1 {
		return "", "", "", 0, fmt.Errorf("search accepts at most one query argument")
	}
	if len(args) == 1 {
		query = strings.TrimSpace(args[0])
	}

	if address != "" {
		if query != "" {
			return "", "", "", 0, fmt.Errorf("provide either <query> or --address, not both")
		}
		if chain == "" {
			return "", "", "", 0, fmt.Errorf("--chain is required with --address")
		}
		return "", chain, address, limit, nil
	}

	if chain != "" {
		return "", "", "", 0, fmt.Errorf("--chain is only supported with --address")
	}
	if query == "" {
		return "", "", "", 0, fmt.Errorf("search requires either <query> or --address")
	}

	return query, "", "", limit, nil
}

func printAddressSearchDryRun(cfg *config.Config, chain, address string, limit int) error {
	headerKey, _ := cfg.AuthHeader()
	masked := cfg.MaskedKey()
	selector, err := api.ResolveDEXNetworkSelector(chain)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"Accept":     "application/json",
		"User-Agent": userAgent,
	}
	if cfg.APIKey != "" {
		headers[headerKey] = masked
	}

	outs := []dryRunOutput{
		{
			Method: "GET",
			URL:    cfg.BaseURL() + "/v4/dex/pairs/quotes/latest",
			Params: map[string]string{
				"contract_address":  address,
				selector.ParamKey(): selector.ParamValue(),
			},
			Headers: headers,
			Note:    "Pair lookup runs first using a local chain selector.",
		},
	}
	for _, dexSlug := range api.DEXSearchCandidateSlugs(selector.NetworkSlug) {
		outs = append(outs, dryRunOutput{
			Method: "GET",
			URL:    cfg.BaseURL() + "/v4/dex/spot-pairs/latest",
			Params: map[string]string{
				selector.ParamKey(): selector.ParamValue(),
				"dex_slug":          dexSlug,
				"limit":             "100",
				"sort":              "liquidity",
				"sort_dir":          "desc",
				"convert_id":        "2781",
			},
			Headers: headers,
			Note:    "Runtime pages spot-pairs/latest for candidate DEX slugs and filters matching token contracts client-side.",
		})
	}

	return printJSONRaw(outs)
}

func renderSearchTable(results []api.SearchResult) {
	headers := []string{"Kind", "Chain", "ID", "Slug", "Symbol", "Name", "Address", "DEX", "Pair", "Liquidity", "Volume 24h", "Rank"}
	rows := make([][]string, len(results))
	for i, result := range results {
		rows[i] = []string{
			display.SanitizeCell(result.Kind),
			display.SanitizeCell(result.Chain),
			formatSearchID(result.ID),
			display.SanitizeCell(result.Slug),
			display.SanitizeCell(result.Symbol),
			display.SanitizeCell(result.Name),
			display.SanitizeCell(result.Address),
			display.SanitizeCell(result.DEX),
			display.SanitizeCell(result.Pair),
			display.FormatLargeNumber(result.Liquidity, "usd"),
			display.FormatLargeNumber(result.Volume24h, "usd"),
			display.FormatRank(result.Rank),
		}
	}
	display.PrintTable(headers, rows)
}

func formatSearchID(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

func renderResolveCandidates(err *api.ResolverAmbiguityError) {
	warnf("Multiple assets match symbol %q. Rerun with --id or --slug.\n", err.Symbol)
	warnf("  ID  Slug           Symbol  Name            Rank  Active\n")
	for _, candidate := range err.Candidates {
		warnf(
			"  %d  %s  %s  %s  %s  %s\n",
			candidate.ID,
			display.SanitizeCell(candidate.Slug),
			display.SanitizeCell(candidate.Symbol),
			display.SanitizeCell(candidate.Name),
			formatResolvedRank(candidate),
			formatResolvedActive(candidate),
		)
	}
}

func formatResolvedRank(asset api.ResolvedAsset) string {
	if !asset.HasRank {
		return "-"
	}
	return display.FormatRank(asset.Rank)
}

func formatResolvedActive(asset api.ResolvedAsset) string {
	if !asset.HasActive {
		return "-"
	}
	return fmt.Sprintf("%t", asset.IsActive)
}
