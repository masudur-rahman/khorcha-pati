package transaction

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"

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

func (ts *txnService) AddTransaction(txn models.Transaction) (err error) {
	if txn.UserID == 0 {
		return models.ErrInvalidTransaction{Reason: "userid is required"}
	}
	if txn.SubcategoryID == "" {
		return models.ErrInvalidTransaction{Reason: "subcategory is required"}
	}
	if txn.Amount <= 0 && txn.SubcategoryID != "misc-init" {
		return models.ErrInvalidTransaction{Reason: "amount must be greater than zero"}
	}
	if types, ok := models.SubcategoryTypes[txn.SubcategoryID]; ok {
		if !models.ContainsType(types, txn.Type) {
			return models.ErrInvalidTransaction{Reason: fmt.Sprintf("subcategory %q is not allowed for transaction type %q", txn.SubcategoryID, txn.Type)}
		}
	}
	if txn.CreatedAt == 0 {
		txn.CreatedAt = time.Now().Unix()
	}
	if err = ts.resolveContact(&txn); err != nil {
		return err
	}

	uow, err := ts.uow.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = uow.Rollback()
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
	err = ts.txnRepo.WithUnitOfWork(uow).AddTransaction(txn)
	return err
}

// GetTransactionByID fetches a transaction and verifies ownership.
func (ts *txnService) GetTransactionByID(userID, txnID int64) (*models.Transaction, error) {
	txn, err := ts.txnRepo.GetTransactionByID(txnID)
	if err != nil {
		return nil, err
	}
	if txn.UserID != userID {
		return nil, models.ErrInvalidTransaction{Reason: "transaction does not belong to user"}
	}
	if txn.DeletedAt != 0 {
		return nil, models.ErrInvalidTransaction{Reason: "transaction has been deleted"}
	}
	return txn, nil
}

// UpdateTransaction reverses the old transaction's balance impact and applies the new one.
func (ts *txnService) UpdateTransaction(userID, txnID int64, updated models.Transaction) (err error) {
	old, err := ts.GetTransactionByID(userID, txnID)
	if err != nil {
		return err
	}

	uow, err := ts.uow.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = uow.Rollback()
			return
		}
		err = uow.Commit()
	}()

	if err = ts.reverseBalances(uow, *old); err != nil {
		return err
	}

	updated.ID = txnID
	updated.UserID = userID
	updated.CreatedAt = old.CreatedAt
	if err = ts.applyBalances(uow, updated); err != nil {
		return err
	}

	err = ts.txnRepo.WithUnitOfWork(uow).UpdateTransaction(txnID, &updated)
	return err
}

// DeleteTransaction soft-deletes a transaction and reverses its balance impact.
func (ts *txnService) DeleteTransaction(userID, txnID int64) (err error) {
	txn, err := ts.GetTransactionByID(userID, txnID)
	if err != nil {
		return err
	}

	uow, err := ts.uow.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = uow.Rollback()
			return
		}
		err = uow.Commit()
	}()

	if err = ts.reverseBalances(uow, *txn); err != nil {
		return err
	}

	err = ts.txnRepo.WithUnitOfWork(uow).SoftDeleteTransaction(txnID, time.Now().Unix())
	return err
}

// Undo soft-deletes the last active transaction and reverses wallet/contact balances.
func (ts *txnService) Undo(userID int64) (txn *models.Transaction, err error) {
	txn, err = ts.txnRepo.GetLastActiveTransaction(userID)
	if err != nil {
		return nil, fmt.Errorf("nothing to undo: %w", err)
	}

	uow, err := ts.uow.Begin(context.Background())
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

	err = ts.txnRepo.WithUnitOfWork(uow).SoftDeleteTransaction(txn.ID, time.Now().Unix())
	return txn, err
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

// applyBalances applies the wallet/contact balance changes for a transaction.
func (ts *txnService) applyBalances(uow styx.UnitOfWork, txn models.Transaction) error {
	walletRepo := ts.walletRepo.WithUnitOfWork(uow)

	switch txn.Type {
	case models.ExpenseTransaction:
		switch txn.SubcategoryID {
		case models.LoanRepaymentSubID, models.BorrowReturnSubID, models.LendSubID:
			if err := ts.updateContactBalance(uow, txn, txn.Amount); err != nil {
				return err
			}
		}
		return walletRepo.UpdateWalletBalance(txn.UserID, txn.SrcID, -txn.Amount)
	case models.IncomeTransaction:
		switch txn.SubcategoryID {
		case models.BorrowSubID, models.LendRecoverySubID, models.LoanReceivedSubID:
			if err := ts.updateContactBalance(uow, txn, -txn.Amount); err != nil {
				return err
			}
		}
		return walletRepo.UpdateWalletBalance(txn.UserID, txn.DstID, txn.Amount)
	case models.TransferTransaction:
		if err := walletRepo.UpdateWalletBalance(txn.UserID, txn.SrcID, -txn.Amount); err != nil {
			return err
		}
		return walletRepo.UpdateWalletBalance(txn.UserID, txn.DstID, txn.Amount)
	}
	return nil
}

// resolveContact normalizes ContactName to a known contact's nickname. When no
// such contact exists the person is recorded in remarks and ContactName is
// cleared — mirroring the bot parser, where unknown people are notes, not
// contacts (so debt balance updates only touch real contacts).
func (ts *txnService) resolveContact(txn *models.Transaction) error {
	if txn.ContactName == "" || txn.Type == models.TransferTransaction {
		return nil
	}
	contact, err := ts.contactRepo.GetContactByName(txn.UserID, txn.ContactName)
	if err == nil {
		if contact.NickName != "" {
			txn.ContactName = contact.NickName
		}
		return nil
	}
	if !models.IsErrNotFound(err) {
		return err
	}
	txn.Remarks = appendPersonNote(txn.Remarks, *txn)
	txn.ContactName = ""
	return nil
}

// appendPersonNote records an unknown person in remarks, e.g. "to meem" or,
// for a debt transaction, "to meem [Person: meem]".
func appendPersonNote(remarks string, txn models.Transaction) string {
	prefix := "to"
	if txn.Type == models.IncomeTransaction {
		prefix = "from"
	}
	note := fmt.Sprintf("%s %s", prefix, txn.ContactName)
	if isDebtSubcategory(txn.SubcategoryID) {
		note = fmt.Sprintf("%s [Person: %s]", note, txn.ContactName)
	}
	if remarks == "" {
		return note
	}
	if strings.Contains(strings.ToLower(remarks), strings.ToLower(note)) {
		return remarks
	}
	return remarks + " " + note
}

// isDebtSubcategory reports whether a subcategory settles a debt with a person.
func isDebtSubcategory(subID string) bool {
	switch subID {
	case models.BorrowSubID, models.BorrowReturnSubID, models.LendSubID, models.LendRecoverySubID:
		return true
	}
	return false
}

func (ts *txnService) updateContactBalance(uow styx.UnitOfWork, txn models.Transaction, amount float64) error {
	if txn.ContactName == "" {
		return nil
	}
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

func (ts *txnService) ListTransactionsPaged(q models.TxnListQuery) ([]models.Transaction, int64, error) {
	return ts.txnRepo.ListTransactionsPaged(q)
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
