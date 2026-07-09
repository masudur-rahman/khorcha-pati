package repos

import (
	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/masudur-rahman/styx"
)

type TransactionRepository interface {
	WithUnitOfWork(uow styx.UnitOfWork) TransactionRepository
	AddTransaction(txn models.Transaction) error
	GetTransactionByID(id int64) (*models.Transaction, error)
	GetLastActiveTransaction(userID int64) (*models.Transaction, error)
	SoftDeleteTransaction(txnID int64, deletedAt int64) error
	UpdateTransaction(id int64, txn *models.Transaction) error
	ListTransactionsByCategory(userID int64, catID string) ([]models.Transaction, error)
	ListTransactions(filter models.Transaction) ([]models.Transaction, error)
	ListTransactionsByTime(userID int64, txnType models.TransactionType, startTime, endTime int64) ([]models.Transaction, error)
	UpdateTransactionsWallet(userID int64, oldShortName, newShortName string) error
	UpdateTransactionsContact(userID int64, oldNickName, newNickName string) error

	GetTxnCategoryName(catID string) (string, error)
	ListTxnCategories() ([]models.TxnCategory, error)
	GetTxnSubcategoryName(subcatID string) (string, error)
	ListTxnSubcategories(catID string) ([]models.TxnSubcategory, error)
	UpdateTxnCategories() error
}
