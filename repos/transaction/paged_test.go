package transaction

import (
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedPagedFixtures inserts a known spread of txns for the paged/scoped tests.
func seedPagedFixtures(t *testing.T, env testEnv) {
	t.Helper()
	fixtures := []models.Transaction{
		{UserID: testUserID, Amount: 10, Type: models.ExpenseTransaction, SrcID: "cash", Timestamp: 100},
		{UserID: testUserID, Amount: 20, Type: models.ExpenseTransaction, SrcID: "brac", Timestamp: 200},
		{UserID: testUserID, Amount: 30, Type: models.IncomeTransaction, DstID: "cash", Timestamp: 300},
		{UserID: testUserID, Amount: 40, Type: models.IncomeTransaction, DstID: "cash", ContactName: "karim", Timestamp: 400},
		{UserID: testUserID, Amount: 50, Type: models.TransferTransaction, SrcID: "cash", DstID: "brac", Timestamp: 500},
	}
	for _, f := range fixtures {
		seedTransaction(t, env, f)
	}
	// Another user's txn — must never leak into results.
	seedTransaction(t, env, models.Transaction{UserID: testUserID + 1, Amount: 999, Type: models.ExpenseTransaction, SrcID: "cash", Timestamp: 600})
}

func TestListTransactionsPaged_pagesAndOrdersDesc(t *testing.T) {
	env := setupTxnRepo(t)
	seedPagedFixtures(t, env)

	// Page 1, limit 2 → newest two by timestamp desc (500, 400).
	txns, total, err := env.repo.ListTransactionsPaged(models.TxnListQuery{UserID: testUserID, Page: 1, Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	require.Len(t, txns, 2)
	assert.Equal(t, int64(500), txns[0].Timestamp)
	assert.Equal(t, int64(400), txns[1].Timestamp)

	// Page 3, limit 2 → the leftover single oldest (100).
	txns, total, err = env.repo.ListTransactionsPaged(models.TxnListQuery{UserID: testUserID, Page: 3, Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	require.Len(t, txns, 1)
	assert.Equal(t, int64(100), txns[0].Timestamp)
}

func TestListTransactionsPaged_noLimitReturnsAll(t *testing.T) {
	env := setupTxnRepo(t)
	seedPagedFixtures(t, env)

	txns, total, err := env.repo.ListTransactionsPaged(models.TxnListQuery{UserID: testUserID})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, txns, 5)
}

func TestListTransactionsPaged_walletScopeMatchesSrcOrDst(t *testing.T) {
	env := setupTxnRepo(t)
	seedPagedFixtures(t, env)

	// "cash" appears as src (100, 500) and dst (300, 400) → 4 txns.
	txns, total, err := env.repo.ListTransactionsPaged(models.TxnListQuery{UserID: testUserID, Wallet: "cash"})
	require.NoError(t, err)
	assert.Equal(t, int64(4), total)
	require.Len(t, txns, 4)
	for _, tx := range txns {
		assert.True(t, tx.SrcID == "cash" || tx.DstID == "cash")
	}
}

func TestListTransactionsPaged_contactAndTypeScope(t *testing.T) {
	env := setupTxnRepo(t)
	seedPagedFixtures(t, env)

	txns, total, err := env.repo.ListTransactionsPaged(models.TxnListQuery{UserID: testUserID, Contact: "karim"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, txns, 1)
	assert.Equal(t, int64(400), txns[0].Timestamp)

	txns, total, err = env.repo.ListTransactionsPaged(models.TxnListQuery{UserID: testUserID, Type: models.IncomeTransaction})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, txns, 2)
}
