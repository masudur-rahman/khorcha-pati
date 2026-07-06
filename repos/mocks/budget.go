package mocks

import (
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"

	"github.com/stretchr/testify/mock"
)

// BudgetRepo is a mock for repos.BudgetRepository.
type BudgetRepo struct {
	mock.Mock
}

var _ repos.BudgetRepository = &BudgetRepo{}

func (m *BudgetRepo) GetBudget(userID int64, categoryID string) (*models.Budget, error) {
	args := m.Called(userID, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Budget), args.Error(1)
}

func (m *BudgetRepo) ListBudgets(userID int64) ([]models.Budget, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Budget), args.Error(1)
}

func (m *BudgetRepo) UpsertBudget(budget *models.Budget) error {
	return m.Called(budget).Error(0)
}

func (m *BudgetRepo) DeleteBudget(userID int64, categoryID string) error {
	return m.Called(userID, categoryID).Error(0)
}
