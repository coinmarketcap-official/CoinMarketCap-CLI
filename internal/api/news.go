package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// NewsLatest fetches the latest content feed from CMC.
// https://coinmarketcap.com/api/documentation/v1/#operation/getV1ContentLatest
func (c *Client) NewsLatest(ctx context.Context, start, limit int, language, newsType string) ([]NewsArticle, error) {
	params := url.Values{
		"start":     {fmt.Sprintf("%d", start)},
		"limit":     {fmt.Sprintf("%d", limit)},
		"language":  {language},
		"news_type": {newsType},
	}
	var result NewsLatestResponse
	if err := c.get(ctx, "/v1/content/latest?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	for i := range result.Data {
		result.Data[i].Language = strings.TrimSpace(result.Data[i].Language)
	}
	return result.Data, nil
}
