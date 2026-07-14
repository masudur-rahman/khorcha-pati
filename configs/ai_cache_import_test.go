package configs

import (
	"context"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportAICacheEntries_addsNewAndSkips(t *testing.T) {
	setupAICacheTest(t)
	_, err := CreateAICache("lunch", "food-rest", "expense", 0.9)
	require.NoError(t, err)

	entries := []models.AICache{
		{InputText: "lunch", SubcategoryID: "food-groc", Intent: "expense", Confidence: 1.0},    // conflict
		{InputText: "bus fare", SubcategoryID: "trans-bus", Intent: "expense", Confidence: 1.0}, // new
		{InputText: "  ", SubcategoryID: "x", Intent: "expense", Confidence: 1.0},               // invalid (blank)
	}
	s, err := ImportAICacheEntries(entries, ImportSkip)
	require.NoError(t, err)
	assert.Equal(t, 1, s.Imported)
	assert.Equal(t, 1, s.Skipped)
	assert.Equal(t, 1, s.Invalid)
	assert.Equal(t, 0, s.Overwritten)

	// conflict kept the original classification
	existing, found, err := getAICacheByInputText(testCtx(), "lunch")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "food-rest", existing.SubcategoryID)
}

func TestImportAICacheEntries_overwrite(t *testing.T) {
	setupAICacheTest(t)
	_, err := CreateAICache("lunch", "food-rest", "expense", 0.9)
	require.NoError(t, err)

	s, err := ImportAICacheEntries([]models.AICache{
		{InputText: "lunch", SubcategoryID: "food-groc", Intent: "expense", Confidence: 0.5},
	}, ImportOverwrite)
	require.NoError(t, err)
	assert.Equal(t, 1, s.Overwritten)

	existing, _, _ := getAICacheByInputText(testCtx(), "lunch")
	assert.Equal(t, "food-groc", existing.SubcategoryID)
	assert.Equal(t, 0.5, existing.Confidence)
}

func TestImportAICacheEntries_higherConfidence(t *testing.T) {
	setupAICacheTest(t)
	_, err := CreateAICache("lunch", "food-rest", "expense", 0.6)
	require.NoError(t, err)

	// lower confidence → skipped
	s, err := ImportAICacheEntries([]models.AICache{
		{InputText: "lunch", SubcategoryID: "food-groc", Intent: "expense", Confidence: 0.5},
	}, ImportHigherConfidence)
	require.NoError(t, err)
	assert.Equal(t, 1, s.Skipped)
	existing, _, _ := getAICacheByInputText(testCtx(), "lunch")
	assert.Equal(t, "food-rest", existing.SubcategoryID)

	// higher confidence → overwritten
	s, err = ImportAICacheEntries([]models.AICache{
		{InputText: "lunch", SubcategoryID: "food-groc", Intent: "expense", Confidence: 0.95},
	}, ImportHigherConfidence)
	require.NoError(t, err)
	assert.Equal(t, 1, s.Overwritten)
	existing, _, _ = getAICacheByInputText(testCtx(), "lunch")
	assert.Equal(t, "food-groc", existing.SubcategoryID)
}

func TestExportAICache_returnsAll(t *testing.T) {
	setupAICacheTest(t)
	_, _ = CreateAICache("a", "food-rest", "expense", 1.0)
	_, _ = CreateAICache("b", "trans-bus", "expense", 1.0)

	rows, err := ExportAICache()
	require.NoError(t, err)
	assert.Len(t, rows, 2)
}

func testCtx() context.Context { return context.Background() }
