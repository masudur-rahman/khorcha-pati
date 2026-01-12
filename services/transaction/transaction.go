package transaction

import (
	"fmt"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"

	"github.com/masudur-rahman/styx"
)

type txnService struct {
	uow       styx.UnitOfWork
	acRepo    repos.AccountsRepository
	drCrRepo  repos.DebtorCreditorRepository
	txnRepo   repos.TransactionRepository
	eventRepo repos.EventRepository
}

func NewTxnService(uow styx.UnitOfWork, acRepo repos.AccountsRepository, drCrRepo repos.DebtorCreditorRepository, txnRepo repos.TransactionRepository, evRepo repos.EventRepository) *txnService {
	return &txnService{
		uow:       uow,
		acRepo:    acRepo,
		drCrRepo:  drCrRepo,
		txnRepo:   txnRepo,
		eventRepo: evRepo,
	}
}

func (ts *txnService) AddTransaction(txn models.Transaction) error {
	if txn.UserID == 0 {
		return fmt.Errorf("userid is required")
	}
	if txn.SubcategoryID == "" {
		return fmt.Errorf("subcategory is required")
	}

	uow, err := ts.uow.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = uow.Rollback()
			return
		}
		err = uow.Commit()
	}()

	switch txn.Type {
	case models.ExpenseTransaction:
		switch txn.SubcategoryID {
		case models.LoanRepaymentSubID, models.BorrowReturnSubID, models.LendSubID:
			if err = ts.updateDebtorCreditorBalance(uow, txn, txn.Amount); err != nil {
				return err
			}
		case models.BorrowSubID, models.LendRecoverySubID, models.LoanReceivedSubID:
			return fmt.Errorf("borrow, lend recovery or loan received type transaction should be under Income type")
		}
		if err = ts.acRepo.WithUnitOfWork(uow).UpdateAccountBalance(txn.UserID, txn.SrcID, -txn.Amount); err != nil {
			return err
		}
	case models.IncomeTransaction:
		switch txn.SubcategoryID {
		case models.BorrowSubID, models.LendRecoverySubID, models.LoanReceivedSubID:
			if err = ts.updateDebtorCreditorBalance(uow, txn, -txn.Amount); err != nil {
				return err
			}
		case models.LoanRepaymentSubID, models.BorrowReturnSubID, models.LendSubID:
			return fmt.Errorf("loan, borrow return or lend type transaction should be under Expense type")
		}
		if err = ts.acRepo.WithUnitOfWork(uow).UpdateAccountBalance(txn.UserID, txn.DstID, txn.Amount); err != nil {
			return err
		}
	case models.TransferTransaction:
		if err = ts.acRepo.WithUnitOfWork(uow).UpdateAccountBalance(txn.UserID, txn.SrcID, -txn.Amount); err != nil {
			return err
		}
		if err = ts.acRepo.WithUnitOfWork(uow).UpdateAccountBalance(txn.UserID, txn.DstID, txn.Amount); err != nil {
			return err
		}
	}
	return ts.txnRepo.WithUnitOfWork(uow).AddTransaction(txn)
}

func (ts *txnService) updateDebtorCreditorBalance(uow styx.UnitOfWork, txn models.Transaction, amount float64) error {
	drcr, err := ts.drCrRepo.WithUnitOfWork(uow).GetDebtorCreditorByName(txn.UserID, txn.DebtorCreditorName)
	if err != nil {
		return err
	}

	return ts.drCrRepo.WithUnitOfWork(uow).UpdateDebtorCreditorBalance(drcr.ID, amount)
}

func (ts *txnService) ListTransactionsByType(userID int64, txnType models.TransactionType) ([]models.Transaction, error) {
	filter := models.Transaction{
		UserID: userID,
		Type:   txnType,
	}
	return ts.txnRepo.ListTransactions(filter)
}

func (ts *txnService) ListTransactions(userID int64) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactions(models.Transaction{UserID: userID})
}

func (ts *txnService) ListTransactionsByCategory(userID int64, catID string) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactionsByCategory(userID, catID)
}

func (ts *txnService) ListTransactionsBySubcategory(userID int64, subcatID string) ([]models.Transaction, error) {
	filter := models.Transaction{
		UserID:        userID,
		SubcategoryID: subcatID,
	}
	return ts.txnRepo.ListTransactions(filter)
}

func (ts *txnService) ListTransactionsByTime(userID int64, txnType models.TransactionType, startTime, endTime int64) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactionsByTime(userID, txnType, startTime, endTime)
}

func (ts *txnService) ListTransactionsBySourceID(userID int64, srcID string) ([]models.Transaction, error) {
	filter := models.Transaction{
		UserID: userID,
		SrcID:  srcID,
	}
	return ts.txnRepo.ListTransactions(filter)
}

func (ts *txnService) ListTransactionsByDestinationID(userID int64, dstID string) ([]models.Transaction, error) {
	filter := models.Transaction{
		UserID: userID,
		DstID:  dstID,
	}
	return ts.txnRepo.ListTransactions(filter)
}

func (ts *txnService) ListTransactionsByDebtorCreditorName(userID int64, drcrName string) ([]models.Transaction, error) {
	filter := models.Transaction{
		UserID:             userID,
		DebtorCreditorName: drcrName,
	}
	return ts.txnRepo.ListTransactions(filter)
}

func (ts *txnService) GetTxnCategoryName(catID string) (string, error) {
	return ts.txnRepo.GetTxnCategoryName(catID)
}

func (ts *txnService) ListTxnCategories() ([]models.TxnCategory, error) {
	return ts.txnRepo.ListTxnCategories()
}

func (ts *txnService) GetTxnSubcategoryName(subcatID string) (string, error) {
	return ts.txnRepo.GetTxnSubcategoryName(subcatID)
}

func (ts *txnService) ListTxnSubcategories(catID string) ([]models.TxnSubcategory, error) {
	return ts.txnRepo.ListTxnSubcategories(catID)
}

func (ts *txnService) UpdateTxnCategories() error {
	return ts.txnRepo.UpdateTxnCategories()
}
