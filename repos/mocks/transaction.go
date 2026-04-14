package mocks

import (
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"

	"github.com/masudur-rahman/styx"

	"github.com/stretchr/testify/mock"
)

// TransactionRepo is a mock for repos.TransactionRepository.
type TransactionRepo struct {
	mock.Mock
}

var _ repos.TransactionRepository = &TransactionRepo{}

func (m *TransactionRepo) WithUnitOfWork(_ styx.UnitOfWork) repos.TransactionRepository {
	return m
}

func (m *TransactionRepo) AddTransaction(txn models.Transaction) error {
	return m.Called(txn).Error(0)
}

func (m *TransactionRepo) GetTransactionByID(id int64) (*models.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *TransactionRepo) UpdateTransaction(id int64, txn *models.Transaction) error {
	return m.Called(id, txn).Error(0)
}

func (m *TransactionRepo) GetLastActiveTransaction(userID int64) (*models.Transaction, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *TransactionRepo) SoftDeleteTransaction(txnID int64, deletedAt int64) error {
	return m.Called(txnID, deletedAt).Error(0)
}

func (m *TransactionRepo) ListTransactionsByCategory(userID int64, catID string) ([]models.Transaction, error) {
	args := m.Called(userID, catID)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *TransactionRepo) ListTransactions(filter models.Transaction) ([]models.Transaction, error) {
	args := m.Called(filter)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *TransactionRepo) ListTransactionsByTime(userID int64, txnType models.TransactionType, start, end int64) ([]models.Transaction, error) {
	args := m.Called(userID, txnType, start, end)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *TransactionRepo) GetTxnCategoryName(catID string) (string, error) {
	args := m.Called(catID)
	return args.String(0), args.Error(1)
}

func (m *TransactionRepo) ListTxnCategories() ([]models.TxnCategory, error) {
	args := m.Called()
	return args.Get(0).([]models.TxnCategory), args.Error(1)
}

func (m *TransactionRepo) GetTxnSubcategoryName(subcatID string) (string, error) {
	args := m.Called(subcatID)
	return args.String(0), args.Error(1)
}

func (m *TransactionRepo) ListTxnSubcategories(catID string) ([]models.TxnSubcategory, error) {
	args := m.Called(catID)
	return args.Get(0).([]models.TxnSubcategory), args.Error(1)
}

func (m *TransactionRepo) UpdateTxnCategories() error {
	return m.Called().Error(0)
}
