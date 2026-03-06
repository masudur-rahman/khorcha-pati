package transaction

import (
	"fmt"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"

	"github.com/masudur-rahman/styx"
)

type txnService struct {
	uow         styx.UnitOfWork
	walletRepo  repos.WalletRepository
	contactRepo repos.ContactRepository
	txnRepo     repos.TransactionRepository
	eventRepo   repos.EventRepository
}

func NewTxnService(uow styx.UnitOfWork, walletRepo repos.WalletRepository, contactRepo repos.ContactRepository, txnRepo repos.TransactionRepository, evRepo repos.EventRepository) *txnService {
	return &txnService{
		uow:         uow,
		walletRepo:  walletRepo,
		contactRepo: contactRepo,
		txnRepo:     txnRepo,
		eventRepo:   evRepo,
	}
}

func (ts *txnService) AddTransaction(txn models.Transaction) error {
	if txn.UserID == 0 {
		return fmt.Errorf("userid is required")
	}
	if txn.SubcategoryID == "" {
		return fmt.Errorf("subcategory is required")
	}
	if txn.CreatedAt == 0 {
		txn.CreatedAt = time.Now().Unix()
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
			if err = ts.updateContactBalance(uow, txn, txn.Amount); err != nil {
				return err
			}
		case models.BorrowSubID, models.LendRecoverySubID, models.LoanReceivedSubID:
			return fmt.Errorf("borrow, lend recovery or loan received type transaction should be under Income type")
		}
		if err = ts.walletRepo.WithUnitOfWork(uow).UpdateWalletBalance(txn.UserID, txn.SrcID, -txn.Amount); err != nil {
			return err
		}
	case models.IncomeTransaction:
		switch txn.SubcategoryID {
		case models.BorrowSubID, models.LendRecoverySubID, models.LoanReceivedSubID:
			if err = ts.updateContactBalance(uow, txn, -txn.Amount); err != nil {
				return err
			}
		case models.LoanRepaymentSubID, models.BorrowReturnSubID, models.LendSubID:
			return fmt.Errorf("loan, borrow return or lend type transaction should be under Expense type")
		}
		if err = ts.walletRepo.WithUnitOfWork(uow).UpdateWalletBalance(txn.UserID, txn.DstID, txn.Amount); err != nil {
			return err
		}
	case models.TransferTransaction:
		if err = ts.walletRepo.WithUnitOfWork(uow).UpdateWalletBalance(txn.UserID, txn.SrcID, -txn.Amount); err != nil {
			return err
		}
		if err = ts.walletRepo.WithUnitOfWork(uow).UpdateWalletBalance(txn.UserID, txn.DstID, txn.Amount); err != nil {
			return err
		}
	}
	return ts.txnRepo.WithUnitOfWork(uow).AddTransaction(txn)
}

// Undo soft-deletes the last active transaction and reverses wallet/contact balances.
func (ts *txnService) Undo(userID int64) (*models.Transaction, error) {
	txn, err := ts.txnRepo.GetLastActiveTransaction(userID)
	if err != nil {
		return nil, fmt.Errorf("nothing to undo: %w", err)
	}

	uow, err := ts.uow.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = uow.Rollback()
			return
		}
		err = uow.Commit()
	}()

	if err = ts.reverseBalances(uow, *txn); err != nil {
		return nil, err
	}

	if err = ts.txnRepo.WithUnitOfWork(uow).SoftDeleteTransaction(txn.ID, time.Now().Unix()); err != nil {
		return nil, err
	}

	return txn, nil
}

func (ts *txnService) reverseBalances(uow styx.UnitOfWork, txn models.Transaction) error {
	walletRepo := ts.walletRepo.WithUnitOfWork(uow)

	switch txn.Type {
	case models.ExpenseTransaction:
		switch txn.SubcategoryID {
		case models.LoanRepaymentSubID, models.BorrowReturnSubID, models.LendSubID:
			if err := ts.updateContactBalance(uow, txn, -txn.Amount); err != nil {
				return err
			}
		}
		return walletRepo.UpdateWalletBalance(txn.UserID, txn.SrcID, txn.Amount)
	case models.IncomeTransaction:
		switch txn.SubcategoryID {
		case models.BorrowSubID, models.LendRecoverySubID, models.LoanReceivedSubID:
			if err := ts.updateContactBalance(uow, txn, txn.Amount); err != nil {
				return err
			}
		}
		return walletRepo.UpdateWalletBalance(txn.UserID, txn.DstID, -txn.Amount)
	case models.TransferTransaction:
		if err := walletRepo.UpdateWalletBalance(txn.UserID, txn.SrcID, txn.Amount); err != nil {
			return err
		}
		return walletRepo.UpdateWalletBalance(txn.UserID, txn.DstID, -txn.Amount)
	}
	return nil
}

func (ts *txnService) updateContactBalance(uow styx.UnitOfWork, txn models.Transaction, amount float64) error {
	contact, err := ts.contactRepo.WithUnitOfWork(uow).GetContactByName(txn.UserID, txn.ContactName)
	if err != nil {
		return err
	}
	return ts.contactRepo.WithUnitOfWork(uow).UpdateContactBalance(contact.ID, amount)
}

func (ts *txnService) ListTransactionsByType(userID int64, txnType models.TransactionType) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactions(models.Transaction{UserID: userID, Type: txnType})
}

func (ts *txnService) ListTransactions(userID int64) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactions(models.Transaction{UserID: userID})
}

func (ts *txnService) ListTransactionsByCategory(userID int64, catID string) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactionsByCategory(userID, catID)
}

func (ts *txnService) ListTransactionsBySubcategory(userID int64, subcatID string) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactions(models.Transaction{UserID: userID, SubcategoryID: subcatID})
}

func (ts *txnService) ListTransactionsByTime(userID int64, txnType models.TransactionType, startTime, endTime int64) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactionsByTime(userID, txnType, startTime, endTime)
}

func (ts *txnService) ListTransactionsBySourceID(userID int64, srcID string) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactions(models.Transaction{UserID: userID, SrcID: srcID})
}

func (ts *txnService) ListTransactionsByDestinationID(userID int64, dstID string) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactions(models.Transaction{UserID: userID, DstID: dstID})
}

func (ts *txnService) ListTransactionsByContactName(userID int64, name string) ([]models.Transaction, error) {
	return ts.txnRepo.ListTransactions(models.Transaction{UserID: userID, ContactName: name})
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
