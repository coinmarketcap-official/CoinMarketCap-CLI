package api

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const categoryPageLimit = 5000

var categoryIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

// ListingsLatest fetches paginated list of cryptocurrencies with market data (CMC).
// https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyListingsLatest
func (c *Client) ListingsLatest(ctx context.Context, start, limit int, convert, sort, sortDir string) ([]ListingCoin, error) {
	return c.ListingsLatestWithCategory(ctx, start, limit, convert, sort, sortDir, "")
}

func (c *Client) ListingsLatestWithCategory(ctx context.Context, start, limit int, convert, sort, sortDir, category string) ([]ListingCoin, error) {
	category = strings.TrimSpace(category)
	if category != "" {
		categoryID, err := c.ResolveCategoryID(ctx, category)
		if err != nil {
			return nil, err
		}
		return c.CategoryCoins(ctx, categoryID, start, limit, convert, sort, sortDir)
	}

	params := url.Values{
		"start":   {fmt.Sprintf("%d", start)},
		"limit":   {fmt.Sprintf("%d", limit)},
		"convert": {convert},
	}
	if sort != "" {
		params.Set("sort", sort)
	}
	if sortDir != "" {
		params.Set("sort_dir", sortDir)
	}
	if category != "" {
		params.Set("category", category)
	}
	var result ListingsLatestResponse
	err := c.get(ctx, "/v1/cryptocurrency/listings/latest?"+params.Encode(), &result)
	return result.Data, err
}

func (c *Client) CategoriesList(ctx context.Context, start, limit int) ([]CategorySummary, error) {
	params := url.Values{
		"start": {fmt.Sprintf("%d", start)},
		"limit": {fmt.Sprintf("%d", limit)},
	}
	var result CategoriesResponse
	if err := c.get(ctx, "/v1/cryptocurrency/categories?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) CategoryCoins(ctx context.Context, categoryID string, start, limit int, convert, sort, sortDir string) ([]ListingCoin, error) {
	params := url.Values{
		"id":      {categoryID},
		"start":   {fmt.Sprintf("%d", start)},
		"limit":   {fmt.Sprintf("%d", limit)},
		"convert": {convert},
	}
	if sort != "" {
		params.Set("sort", sort)
	}
	if sortDir != "" {
		params.Set("sort_dir", sortDir)
	}

	var result CategoryResponse
	if err := c.get(ctx, "/v1/cryptocurrency/category?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	return result.Data.Coins, nil
}

func (c *Client) ResolveCategoryID(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrInvalidInput
	}
	if categoryIDPattern.MatchString(token) {
		return token, nil
	}

	normalized := normalizeCategoryToken(token)
	start := 1
	for {
		categories, err := c.CategoriesList(ctx, start, categoryPageLimit)
		if err != nil {
			return "", err
		}
		if len(categories) == 0 {
			break
		}
		for _, category := range categories {
			if categoryMatchesToken(category, token, normalized) {
				return category.ID, nil
			}
		}
		if len(categories) < categoryPageLimit {
			break
		}
		start += categoryPageLimit
	}

	return "", fmt.Errorf("category %q not found (%w)", token, ErrAssetNotFound)
}

func categoryMatchesToken(category CategorySummary, token, normalized string) bool {
	candidates := []string{
		category.ID,
		category.Name,
		category.Title,
		category.Slug,
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if strings.EqualFold(candidate, token) {
			return true
		}
		if normalizeCategoryToken(candidate) == normalized {
			return true
		}
	}
	return false
}

func normalizeCategoryToken(value string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if b.Len() > 0 && !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// QuotesLatestByID fetches current price quotes by CMC IDs (e.g. "1", "1027").
// https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyQuotesLatest
func (c *Client) QuotesLatestByID(ctx context.Context, ids []string, convert string) (map[string]QuoteCoin, error) {
	params := url.Values{
		"id":      {strings.Join(ids, ",")},
		"convert": {convert},
	}
	var result QuotesLatestResponse
	err := c.get(ctx, "/v2/cryptocurrency/quotes/latest?"+params.Encode(), &result)
	return result.Data, err
}

// QuotesLatestBySlug fetches current price quotes by CMC slugs (e.g. "bitcoin", "ethereum").
// https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyQuotesLatest
func (c *Client) QuotesLatestBySlug(ctx context.Context, slugs []string, convert string) (map[string]QuoteCoin, error) {
	params := url.Values{
		"slug":    {strings.Join(slugs, ",")},
		"convert": {convert},
	}
	var result QuotesLatestResponse
	err := c.get(ctx, "/v2/cryptocurrency/quotes/latest?"+params.Encode(), &result)
	return result.Data, err
}

// QuotesLatestBySymbol fetches current price quotes by symbols (e.g. "BTC", "ETH").
// https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyQuotesLatest
func (c *Client) QuotesLatestBySymbol(ctx context.Context, symbols []string, convert string) (map[string]QuoteCoin, error) {
	params := url.Values{
		"symbol":  {strings.Join(symbols, ",")},
		"convert": {convert},
	}
	var result QuotesLatestResponse
	err := c.get(ctx, "/v2/cryptocurrency/quotes/latest?"+params.Encode(), &result)
	return result.Data, err
}

// TrendingGainersLosers fetches top gaining and losing trending cryptocurrencies (CMC).
// https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyTrendingGainersLosers
func (c *Client) TrendingGainersLosers(ctx context.Context, start, limit int, timePeriod, convert string) ([]TrendingCoin2, error) {
	params := url.Values{
		"start":       {fmt.Sprintf("%d", start)},
		"limit":       {fmt.Sprintf("%d", limit)},
		"time_period": {timePeriod},
		"convert":     {convert},
	}
	var result TrendingGainersLosersResponse
	err := c.get(ctx, "/v1/cryptocurrency/trending/gainers-losers?"+params.Encode(), &result)
	return result.Data, err
}

// TrendingLatest fetches the latest trending cryptocurrencies (CMC).
// https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyTrendingLatest
func (c *Client) TrendingLatest(ctx context.Context, start, limit int, convert string) ([]ListingCoin, error) {
	params := url.Values{
		"start":   {fmt.Sprintf("%d", start)},
		"limit":   {fmt.Sprintf("%d", limit)},
		"convert": {convert},
	}
	var result ListingsLatestResponse
	err := c.get(ctx, "/v1/cryptocurrency/trending/latest?"+params.Encode(), &result)
	return result.Data, err
}
