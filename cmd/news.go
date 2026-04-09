package cmd

import (
	"fmt"
	"strings"

	"github.com/openCMC/CoinMarketCap-CLI/internal/display"

	"github.com/spf13/cobra"
)

var newsCmd = &cobra.Command{
	Use:   "news",
	Short: "Show latest CMC news content",
	Long: `Fetch CoinMarketCap news from the provider-backed latest content endpoint.

This command supports pagination and language/news-type filters.`,
Example: `  cmc news
  cmc news --limit 5
  cmc news --start 21 --language en --news-type news
  cmc news --dry-run -o json`,
	RunE: runNews,
}

func init() {
	newsCmd.Flags().Int("start", 1, "Offset for pagination")
	newsCmd.Flags().Int("limit", 20, "Number of news items to fetch (1-100)")
	newsCmd.Flags().String("language", "en", "Content language")
	newsCmd.Flags().String("news-type", "news", "News type filter")
	addOutputFlag(newsCmd)
	addDryRunFlag(newsCmd)
	rootCmd.AddCommand(newsCmd)
}

type newsItem struct {
	Title      string   `json:"title"`
	Subtitle   string   `json:"subtitle"`
	SourceName string   `json:"source_name"`
	SourceURL  string   `json:"source_url"`
	ReleasedAt string   `json:"released_at"`
	CreatedAt  string   `json:"created_at"`
	Assets     []string `json:"assets"`
	Type       string   `json:"type"`
	NewsType   string   `json:"news_type"`
	Cover      string   `json:"cover"`
	Language   string   `json:"language"`
}

func runNews(cmd *cobra.Command, args []string) error {
	start, _ := cmd.Flags().GetInt("start")
	limit, _ := cmd.Flags().GetInt("limit")
	language, _ := cmd.Flags().GetString("language")
	newsType, _ := cmd.Flags().GetString("news-type")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}
	if start < 1 {
		return fmt.Errorf("--start must be at least 1")
	}
	if limit < 1 || limit > 100 {
		return fmt.Errorf("--limit must be between 1 and 100")
	}
	language = strings.TrimSpace(language)
	if language == "" {
		language = "en"
	}
	newsType = strings.TrimSpace(newsType)
	if newsType == "" {
		newsType = "news"
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if isDryRun(cmd) {
		params := map[string]string{
			"start":     fmt.Sprintf("%d", start),
			"limit":     fmt.Sprintf("%d", limit),
			"language":  language,
			"news_type": newsType,
		}
		return printDryRun(cfg, "news", "/v1/content/latest", params, nil)
	}

	client := newAPIClient(cfg)
	ctx := cmd.Context()

	articles, err := client.NewsLatest(ctx, start, limit, language, newsType)
	if err != nil {
		return err
	}

	items := make([]newsItem, 0, len(articles))
	for _, article := range articles {
		symbols := make([]string, 0, len(article.Assets))
		for _, asset := range article.Assets {
			if sym := strings.TrimSpace(asset.Symbol); sym != "" {
				symbols = append(symbols, sym)
			}
		}
		items = append(items, newsItem{
			Title:      article.Title,
			Subtitle:   article.Subtitle,
			SourceName: article.SourceName,
			SourceURL:  article.SourceURL,
			ReleasedAt: article.ReleasedAt,
			CreatedAt:  article.CreatedAt,
			Assets:     symbols,
			Type:       article.Type,
			NewsType:   article.NewsType,
			Cover:      article.Cover,
			Language:   article.Language,
		})
	}

	if jsonOut {
		return printJSONRaw(items)
	}

	fmt.Printf("CMC News (start %d, limit %d, language %s, type %s)\n\n", start, limit, language, newsType)
	headers := []string{"Title", "Source", "Released At", "Assets"}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{
			display.SanitizeCell(item.Title),
			display.SanitizeCell(item.SourceName),
			display.SanitizeCell(item.ReleasedAt),
			display.SanitizeCell(strings.Join(item.Assets, ", ")),
		})
	}
	display.PrintTable(headers, rows)
	return nil
}
