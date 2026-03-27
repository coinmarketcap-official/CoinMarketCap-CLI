package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) InfoByID(ctx context.Context, id string) (*CoinInfo, error) {
	infos, err := c.infoByQuery(ctx, url.Values{"id": {id}})
	if err != nil {
		return nil, err
	}
	return firstCoinInfo(infos)
}

func (c *Client) InfoBySlug(ctx context.Context, slug string) (*CoinInfo, error) {
	infos, err := c.infoByQuery(ctx, url.Values{"slug": {slug}})
	if err != nil {
		return nil, err
	}
	return firstCoinInfo(infos)
}

func (c *Client) InfoByIDs(ctx context.Context, ids []string) (map[string]CoinInfo, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("info not found (%w)", ErrAssetNotFound)
	}
	return c.infoByQuery(ctx, url.Values{"id": {strings.Join(ids, ",")}})
}

func (c *Client) infoByQuery(ctx context.Context, params url.Values) (map[string]CoinInfo, error) {
	var result InfoResponse
	if err := c.get(ctx, "/v2/cryptocurrency/info?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("info not found (%w)", ErrAssetNotFound)
	}
	return result.Data, nil
}

func firstCoinInfo(infos map[string]CoinInfo) (*CoinInfo, error) {
	for _, info := range infos {
		infoCopy := info
		return &infoCopy, nil
	}
	return nil, fmt.Errorf("info not found (%w)", ErrAssetNotFound)
}
