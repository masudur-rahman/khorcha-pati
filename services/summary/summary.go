package summary

import (
	"fmt"
	"strings"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"
	"github.com/masudur-rahman/expense-tracker-bot/services"
)

type summaryService struct {
	txnRepo    repos.TransactionRepository
	walletRepo repos.WalletRepository
	budgetRepo repos.BudgetRepository
}

// NewSummaryService creates a new SummaryService.
func NewSummaryService(
	txnRepo repos.TransactionRepository,
	walletRepo repos.WalletRepository,
	budgetRepo repos.BudgetRepository,
) services.SummaryService {
	return &summaryService{
		txnRepo:    txnRepo,
		walletRepo: walletRepo,
		budgetRepo: budgetRepo,
	}
}

func (s *summaryService) GetMonthlyOverview(userID int64, year, month int) (*models.MonthlyOverview, error) {
	start, end := monthRange(year, month)

	txns, err := s.txnRepo.ListTransactionsByTime(userID, "", start, end)
	if err != nil {
		return nil, err
	}

	var income, expense float64
	for _, t := range txns {
		switch t.Type {
		case models.ExpenseTransaction:
			expense += t.Amount
		case models.IncomeTransaction:
			income += t.Amount
		}
	}

	wallets, err := s.walletRepo.ListWallets(userID)
	if err != nil {
		return nil, err
	}
	var totalBalance float64
	for _, w := range wallets {
		totalBalance += w.Balance
	}

	budgetUsage := computeBudgetUsage(s.budgetRepo, userID, txns)

	return &models.MonthlyOverview{
		TotalBalance: totalBalance,
		MonthIncome:  income,
		MonthExpense: expense,
		BudgetUsage:  budgetUsage,
	}, nil
}

func (s *summaryService) GetExpenseByCategory(userID int64, year, month int) ([]models.CategorySpend, error) {
	start, end := monthRange(year, month)

	txns, err := s.txnRepo.ListTransactionsByTime(userID, models.ExpenseTransaction, start, end)
	if err != nil {
		return nil, err
	}

	catTotals := make(map[string]float64)
	var total float64
	for _, t := range txns {
		catID := extractCategoryID(t.SubcategoryID)
		catTotals[catID] += t.Amount
		total += t.Amount
	}

	result := make([]models.CategorySpend, 0, len(catTotals))
	for catID, amount := range catTotals {
		pct := 0.0
		if total > 0 {
			pct = (amount / total) * 100
		}
		result = append(result, models.CategorySpend{
			CategoryID:   catID,
			CategoryName: categoryName(catID),
			Amount:       amount,
			Percent:      pct,
		})
	}
	return result, nil
}

func (s *summaryService) GetIncomeVsExpense(userID int64, months int) ([]models.MonthlyComparison, error) {
	now := time.Now()
	result := make([]models.MonthlyComparison, 0, months)

	for i := months - 1; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		start, end := monthRange(t.Year(), int(t.Month()))

		txns, err := s.txnRepo.ListTransactionsByTime(userID, "", start, end)
		if err != nil {
			return nil, err
		}

		var income, expense float64
		for _, txn := range txns {
			switch txn.Type {
			case models.IncomeTransaction:
				income += txn.Amount
			case models.ExpenseTransaction:
				expense += txn.Amount
			}
		}

		result = append(result, models.MonthlyComparison{
			Month:   fmt.Sprintf("%d-%02d", t.Year(), t.Month()),
			Income:  income,
			Expense: expense,
		})
	}
	return result, nil
}

func monthRange(year, month int) (int64, int64) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	return start.Unix(), end.Unix()
}

func extractCategoryID(subcatID string) string {
	parts := strings.SplitN(subcatID, "-", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return subcatID
}

func categoryName(catID string) string {
	if name, ok := models.SubCatToCatNameMap[catID]; ok {
		return name
	}
	return catID
}

func computeBudgetUsage(budgetRepo repos.BudgetRepository, userID int64, txns []models.Transaction) float64 {
	budgets, err := budgetRepo.ListBudgets(userID)
	if err != nil || len(budgets) == 0 {
		return 0
	}

	catSpent := make(map[string]float64)
	var overallSpent float64
	for _, t := range txns {
		if t.Type == models.ExpenseTransaction {
			catID := extractCategoryID(t.SubcategoryID)
			catSpent[catID] += t.Amount
			overallSpent += t.Amount
		}
	}
	catSpent[""] = overallSpent // "" key = overall budget

	var totalBudget, totalSpent float64
	for _, b := range budgets {
		totalBudget += b.Amount
		if spent, ok := catSpent[b.CategoryID]; ok {
			totalSpent += spent
		}
	}

	if totalBudget == 0 {
		return 0
	}
	return (totalSpent / totalBudget) * 100
}
