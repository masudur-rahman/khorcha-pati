package budgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"
	"github.com/masudur-rahman/expense-tracker-bot/services"
)

const defaultAlertAt = 80

type budgetService struct {
	budgetRepo repos.BudgetRepository
	txnRepo    repos.TransactionRepository
}

var _ services.BudgetService = &budgetService{}

// NewBudgetService creates a new budget service.
func NewBudgetService(budgetRepo repos.BudgetRepository, txnRepo repos.TransactionRepository) *budgetService {
	return &budgetService{budgetRepo: budgetRepo, txnRepo: txnRepo}
}

// SetBudget creates or updates a monthly budget for a category (or overall if categoryID is "").
func (s *budgetService) SetBudget(userID int64, categoryID string, amount float64, alertAt int64) error {
	if amount <= 0 {
		return fmt.Errorf("budget amount must be positive")
	}
	if alertAt <= 0 || alertAt > 100 {
		alertAt = defaultAlertAt
	}
	return s.budgetRepo.UpsertBudget(&models.Budget{
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     amount,
		AlertAt:    alertAt,
	})
}

// GetBudgetStatus returns a single budget with current month spent data.
func (s *budgetService) GetBudgetStatus(userID int64, categoryID string) (*models.BudgetStatus, error) {
	budget, err := s.budgetRepo.GetBudget(userID, categoryID)
	if err != nil {
		return nil, err
	}

	spent, err := s.computeSpent(userID, categoryID)
	if err != nil {
		return nil, err
	}

	return buildStatus(*budget, spent, s.resolveName(categoryID)), nil
}

// ListBudgetStatuses returns all budgets for a user with current month spending.
func (s *budgetService) ListBudgetStatuses(userID int64) ([]models.BudgetStatus, error) {
	budgets, err := s.budgetRepo.ListBudgets(userID)
	if err != nil {
		return nil, err
	}
	if len(budgets) == 0 {
		return nil, nil
	}

	spentByCategory, err := s.computeSpentByCategory(userID)
	if err != nil {
		return nil, err
	}

	result := make([]models.BudgetStatus, 0, len(budgets))
	for _, b := range budgets {
		spent := spentByCategory[b.CategoryID]
		result = append(result, *buildStatus(b, spent, s.resolveName(b.CategoryID)))
	}
	return result, nil
}

// DeleteBudget removes a budget.
func (s *budgetService) DeleteBudget(userID int64, categoryID string) error {
	return s.budgetRepo.DeleteBudget(userID, categoryID)
}

// CheckBudgetAlerts checks category + overall budgets and returns triggered alerts.
func (s *budgetService) CheckBudgetAlerts(userID int64, subcategoryID string) ([]models.BudgetAlert, error) {
	categoryID := strings.Split(subcategoryID, "-")[0]

	budgets, err := s.budgetRepo.ListBudgets(userID)
	if err != nil {
		return nil, err
	}
	if len(budgets) == 0 {
		return nil, nil
	}

	spentByCategory, err := s.computeSpentByCategory(userID)
	if err != nil {
		return nil, err
	}

	var alerts []models.BudgetAlert
	for _, b := range budgets {
		if b.CategoryID != categoryID && b.CategoryID != "" {
			continue
		}
		spent := spentByCategory[b.CategoryID]
		if b.Amount <= 0 {
			continue
		}
		pct := spent / b.Amount * 100
		if pct < float64(b.AlertAt) {
			continue
		}
		alerts = append(alerts, models.BudgetAlert{
			CategoryName: s.resolveName(b.CategoryID),
			Spent:        spent,
			Limit:        b.Amount,
			Percent:      pct,
			Exceeded:     pct >= 100,
		})
	}
	return alerts, nil
}

// computeSpent returns total expense amount for a category in the current month.
func (s *budgetService) computeSpent(userID int64, categoryID string) (float64, error) {
	spentMap, err := s.computeSpentByCategory(userID)
	if err != nil {
		return 0, err
	}
	return spentMap[categoryID], nil
}

// computeSpentByCategory fetches current month expenses and groups by category.
func (s *budgetService) computeSpentByCategory(userID int64) (map[string]float64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	txns, err := s.txnRepo.ListTransactionsByTime(userID, models.ExpenseTransaction, startOfMonth.Unix(), now.Unix())
	if err != nil {
		return nil, err
	}

	result := map[string]float64{}
	var total float64
	for _, txn := range txns {
		cat := strings.Split(txn.SubcategoryID, "-")[0]
		result[cat] += txn.Amount
		total += txn.Amount
	}
	result[""] = total // overall
	return result, nil
}

// resolveName returns a display name for a category ID.
func (s *budgetService) resolveName(categoryID string) string {
	if categoryID == "" {
		return "Overall"
	}
	name, err := s.txnRepo.GetTxnCategoryName(categoryID)
	if err != nil {
		return categoryID
	}
	return name
}

func buildStatus(b models.Budget, spent float64, name string) *models.BudgetStatus {
	remaining := b.Amount - spent
	var pct float64
	if b.Amount > 0 {
		pct = spent / b.Amount * 100
	}
	return &models.BudgetStatus{
		Budget:       b,
		CategoryName: name,
		Spent:        spent,
		Remaining:    remaining,
		Percent:      pct,
	}
}
