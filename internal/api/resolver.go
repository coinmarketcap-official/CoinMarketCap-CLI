package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type ResolvedAsset struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Rank        int    `json:"-"`
	IsActive    bool   `json:"-"`
	HasRank     bool   `json:"-"`
	HasActive   bool   `json:"-"`
}

func (a ResolvedAsset) MarshalJSON() ([]byte, error) {
	type alias struct {
		ID       int64  `json:"id"`
		Slug     string `json:"slug"`
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Rank     *int   `json:"rank"`
		IsActive *bool  `json:"is_active"`
	}

	out := alias{
		ID:     a.ID,
		Slug:   a.Slug,
		Symbol: a.Symbol,
		Name:   a.Name,
	}
	if a.HasRank {
		rank := a.Rank
		out.Rank = &rank
	}
	if a.HasActive {
		active := a.IsActive
		out.IsActive = &active
	}
	return json.Marshal(out)
}

type ResolverAmbiguityError struct {
	Symbol     string          `json:"symbol"`
	Candidates []ResolvedAsset `json:"candidates"`
}

func (e *ResolverAmbiguityError) Error() string {
	return fmt.Sprintf("symbol %q is ambiguous; rerun with --id or --slug", e.Symbol)
}

func (e *ResolverAmbiguityError) Is(target error) bool {
	return target == ErrResolverAmbiguous
}

type mapResponse struct {
	Data []struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Slug     string `json:"slug"`
		Rank     int    `json:"rank"`
		IsActive int    `json:"is_active"`
	} `json:"data"`
}

func (c *Client) ResolveByID(ctx context.Context, id string) (ResolvedAsset, error) {
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return ResolvedAsset{}, fmt.Errorf("id must be numeric (%w)", ErrInvalidInput)
	}
	info, err := c.InfoByID(ctx, id)
	if err != nil {
		return ResolvedAsset{}, err
	}
	if info == nil {
		return ResolvedAsset{}, ErrAssetNotFound
	}
	return resolvedAssetFromInfo(*info), nil
}

func (c *Client) ResolveBySlug(ctx context.Context, slug string) (ResolvedAsset, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return ResolvedAsset{}, ErrInvalidInput
	}

	info, err := c.InfoBySlug(ctx, slug)
	if err != nil {
		return ResolvedAsset{}, err
	}
	if info == nil {
		return ResolvedAsset{}, ErrAssetNotFound
	}
	return resolvedAssetFromInfo(*info), nil
}

// ResolveBySymbolOrSlug tries symbol lookup (uppercase) first, then slug (lowercase) if the symbol
// resolves to no assets. Ambiguity from symbol resolution is returned as-is (no slug fallback).
func (c *Client) ResolveBySymbolOrSlug(ctx context.Context, token string) (ResolvedAsset, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return ResolvedAsset{}, ErrInvalidInput
	}

	symbol := strings.ToUpper(token)
	asset, err := c.ResolveBySymbol(ctx, symbol)
	if err == nil {
		return asset, nil
	}
	var ambig *ResolverAmbiguityError
	if errors.As(err, &ambig) {
		return ResolvedAsset{}, err
	}
	if !errors.Is(err, ErrAssetNotFound) {
		return ResolvedAsset{}, err
	}

	slug := strings.ToLower(token)
	return c.ResolveBySlug(ctx, slug)
}

func (c *Client) ResolveBySymbol(ctx context.Context, symbol string) (ResolvedAsset, error) {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return ResolvedAsset{}, ErrInvalidInput
	}

	assets, err := c.resolveMany(ctx, map[string]string{"symbol": symbol})
	if err != nil {
		return ResolvedAsset{}, err
	}
	if len(assets) == 0 {
		return ResolvedAsset{}, ErrAssetNotFound
	}
	if len(assets) == 1 {
		return assets[0], nil
	}

	sortResolvedAssets(assets)
	return ResolvedAsset{}, &ResolverAmbiguityError{Symbol: symbol, Candidates: assets}
}

func (c *Client) ResolveBySymbolCandidates(ctx context.Context, symbol string, limit int) ([]ResolvedAsset, error) {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return nil, ErrInvalidInput
	}

	assets, err := c.resolveMany(ctx, map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}
	if len(assets) == 0 {
		return nil, ErrAssetNotFound
	}

	sortResolvedAssets(assets)
	if limit > 0 && len(assets) > limit {
		assets = assets[:limit]
	}
	return assets, nil
}

func (c *Client) resolveMany(ctx context.Context, params map[string]string) ([]ResolvedAsset, error) {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}

	var resp mapResponse
	if err := c.get(ctx, "/v1/cryptocurrency/map?"+values.Encode(), &resp); err != nil {
		return nil, err
	}

	assets := make([]ResolvedAsset, 0, len(resp.Data))
	for _, item := range resp.Data {
		assets = append(assets, ResolvedAsset{
			ID:        item.ID,
			Slug:      item.Slug,
			Symbol:    item.Symbol,
			Name:      item.Name,
			Rank:      item.Rank,
			IsActive:  item.IsActive == 1,
			HasRank:   true,
			HasActive: true,
		})
	}
	return assets, nil
}

func sortResolvedAssets(assets []ResolvedAsset) {
	sort.SliceStable(assets, func(i, j int) bool {
		if assets[i].Rank == assets[j].Rank {
			return assets[i].ID < assets[j].ID
		}
		if assets[i].Rank == 0 {
			return false
		}
		if assets[j].Rank == 0 {
			return true
		}
		return assets[i].Rank < assets[j].Rank
	})
}

func resolvedAssetFromInfo(info CoinInfo) ResolvedAsset {
	return ResolvedAsset{
		ID:        info.ID,
		Slug:      info.Slug,
		Symbol:    info.Symbol,
		Name:      info.Name,
		Rank:      0,
		IsActive:  false,
		HasRank:   false,
		HasActive: false,
	}
}
