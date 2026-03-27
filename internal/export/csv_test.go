package export

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")

	headers := []string{"Name", "Price"}
	rows := [][]string{
		{"Bitcoin", "42000.00"},
		{"Ethereum", "3000.00"},
	}

	err := ExportCSV(path, headers, rows)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "Name,Price")
	assert.Contains(t, content, "Bitcoin,42000.00")
	assert.Contains(t, content, "Ethereum,3000.00")
}

func TestExportCSVEmptyRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.csv")

	err := ExportCSV(path, []string{"A", "B"}, nil)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "A,B\n", string(data))
}
