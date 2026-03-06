package services

import "github.com/masudur-rahman/expense-tracker-bot/models"

type TransactionService interface {
	AddTransaction(txn models.Transaction) error
	ListTransactions(userID int64) ([]models.Transaction, error)
	ListTransactionsByType(userID int64, txnType models.TransactionType) ([]models.Transaction, error)
	ListTransactionsByCategory(userID int64, catID string) ([]models.Transaction, error)
	ListTransactionsBySubcategory(userID int64, subcatID string) ([]models.Transaction, error)
	ListTransactionsByTime(userID int64, txnType models.TransactionType, startTime, endTime int64) ([]models.Transaction, error)
	ListTransactionsBySourceID(userID int64, srcID string) ([]models.Transaction, error)
	ListTransactionsByDestinationID(userID int64, dstID string) ([]models.Transaction, error)
	ListTransactionsByContactName(userID int64, name string) ([]models.Transaction, error)

	GetTxnCategoryName(catID string) (string, error)
	ListTxnCategories() ([]models.TxnCategory, error)
	GetTxnSubcategoryName(subcatID string) (string, error)
	ListTxnSubcategories(catID string) ([]models.TxnSubcategory, error)
	UpdateTxnCategories() error
}
