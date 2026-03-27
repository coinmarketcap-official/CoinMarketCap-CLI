package cmd

import "fmt"

func validateExactlyOneSelectorFamily(idStr, slugStr, symbolStr string) error {
	count := 0
	for _, v := range []string{idStr, slugStr, symbolStr} {
		if v != "" {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("provide exactly one of --id, --slug, or --symbol")
	}
	return nil
}
