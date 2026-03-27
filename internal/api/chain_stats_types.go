package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type BlockchainStatisticsLatestResponse struct {
	Data map[string]BlockchainStatistics `json:"data"`
}

type BlockchainStatistics struct {
	ID                  int64  `json:"id"`
	Slug                string `json:"slug"`
	Symbol              string `json:"symbol"`
	BlockRewardStatic   string `json:"block_reward_static"`
	ConsensusMechanism  string `json:"consensus_mechanism"`
	Difficulty          string `json:"difficulty"`
	Hashrate24h         string `json:"hashrate_24h"`
	PendingTransactions string `json:"pending_transactions"`
	ReductionRate       string `json:"reduction_rate"`
	TotalBlocks         string `json:"total_blocks"`
	TotalTransactions   string `json:"total_transactions"`
	TPS24h              string `json:"tps_24h"`
	FirstBlockTimestamp string `json:"first_block_timestamp"`
}

func (b *BlockchainStatistics) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw["id"]; ok {
		switch n := v.(type) {
		case float64:
			b.ID = int64(n)
		case string:
			if n == "" {
				b.ID = 0
			} else if parsed, err := strconv.ParseInt(n, 10, 64); err == nil {
				b.ID = parsed
			} else {
				return fmt.Errorf("parse blockchain statistic id: %w", err)
			}
		}
	}

	stringField := func(key string) string {
		v, ok := raw[key]
		if !ok || v == nil {
			return ""
		}
		switch n := v.(type) {
		case string:
			return n
		case float64:
			return strconv.FormatFloat(n, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(n)
		default:
			return fmt.Sprint(n)
		}
	}

	b.Slug = stringField("slug")
	b.Symbol = stringField("symbol")
	b.BlockRewardStatic = stringField("block_reward_static")
	b.ConsensusMechanism = stringField("consensus_mechanism")
	b.Difficulty = stringField("difficulty")
	b.Hashrate24h = stringField("hashrate_24h")
	b.PendingTransactions = stringField("pending_transactions")
	b.ReductionRate = stringField("reduction_rate")
	b.TotalBlocks = stringField("total_blocks")
	b.TotalTransactions = stringField("total_transactions")
	b.TPS24h = stringField("tps_24h")
	b.FirstBlockTimestamp = stringField("first_block_timestamp")
	return nil
}
