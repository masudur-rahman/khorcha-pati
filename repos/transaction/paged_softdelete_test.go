package transaction

import (
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTransactionsPaged_excludesSoftDeleted(t *testing.T) {
	env := setupTxnRepo(t)
	seedTransaction(t, env, models.Transaction{UserID: testUserID, Amount: 10, Type: models.ExpenseTransaction, SrcID: "cash", Timestamp: 100})
	seedTransaction(t, env, models.Transaction{UserID: testUserID, Amount: 20, Type: models.ExpenseTransaction, SrcID: "cash", Timestamp: 200})
	seedTransaction(t, env, models.Transaction{UserID: testUserID, Amount: 30, Type: models.ExpenseTransaction, SrcID: "cash", Timestamp: 300})

	// soft-delete the newest
	all, _, err := env.repo.ListTransactionsPaged(models.TxnListQuery{UserID: testUserID})
	require.NoError(t, err)
	require.Len(t, all, 3)
	require.NoError(t, env.repo.SoftDeleteTransaction(all[0].ID, 999))

	txns, total, err := env.repo.ListTransactionsPaged(models.TxnListQuery{UserID: testUserID})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total, "count must exclude soft-deleted")
	require.Len(t, txns, 2, "data must exclude soft-deleted")
	for _, tx := range txns {
		assert.Zero(t, tx.DeletedAt)
	}
}
