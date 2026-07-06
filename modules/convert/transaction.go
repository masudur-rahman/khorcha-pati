package convert

import (
	"strings"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/models/gqtypes"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"
	"github.com/masudur-rahman/khorcha-pati/services/all"
)

func ToTransactionAPIFormat(txn models.Transaction) gqtypes.Transaction {
	svc := all.GetServices()
	var err error
	var category, subcategory, src, dst, person string
	catID := strings.Split(txn.SubcategoryID, "-")[0]
	if err = cache.FetchDataWithCustomFunc(catID, &category, func() (any, error) {
		return svc.Txn.GetTxnCategoryName(catID)
	}); err != nil {
		category = catID
	}

	if err = cache.FetchDataWithCustomFunc(txn.SubcategoryID, &subcategory, func() (any, error) {
		return svc.Txn.GetTxnSubcategoryName(txn.SubcategoryID)
	}); err != nil {
		subcategory = txn.SubcategoryID
	}

	if txn.SrcID != "" {
		if err = cache.FetchDataWithCustomFunc(txn.SrcID, &src, func() (any, error) {
			ac, err := svc.Wallet.GetWalletByShortName(txn.UserID, txn.SrcID)
			if err != nil {
				return nil, err
			}
			return ac.Name, nil
		}); err != nil {
			src = txn.SrcID
		}
	}

	if txn.DstID != "" {
		if err = cache.FetchDataWithCustomFunc(txn.DstID, &dst, func() (any, error) {
			ac, err := svc.Wallet.GetWalletByShortName(txn.UserID, txn.DstID)
			if err != nil {
				return nil, err
			}
			return ac.Name, nil
		}); err != nil {
			dst = txn.DstID
		}
	}

	if txn.ContactName != "" {
		if err = cache.FetchDataWithCustomFunc(txn.ContactName, &person, func() (any, error) {
			user, err := svc.Contact.GetContactByName(txn.UserID, txn.ContactName)
			if err != nil {
				return nil, err
			}
			return user.FullName, nil
		}); err != nil {
			person = txn.ContactName
		}
	}

	return gqtypes.Transaction{
		Date:        time.Unix(txn.Timestamp, 0),
		Type:        string(txn.Type),
		Amount:      txn.Amount,
		Source:      src,
		Destination: dst,
		Person:      person,
		Category:    category,
		Subcategory: subcategory,
		Remarks:     txn.Remarks,
	}
}

func ToWalletAPIFormat(w models.Wallet) gqtypes.Wallet {
	return gqtypes.Wallet{
		ID:        w.ID,
		Type:      string(w.Type),
		ShortName: w.ShortName,
		Name:      w.Name,
		Balance:   w.Balance,
		CreatedAt: time.Unix(w.CreatedAt, 0),
	}
}

func ToContactAPIFormat(c models.Contacts) gqtypes.Contact {
	return gqtypes.Contact{
		ID:               c.ID,
		NickName:         c.NickName,
		FullName:         c.FullName,
		Email:            c.Email,
		NetBalance:       c.NetBalance,
		LastTxnTimestamp: c.LastTxnTimestamp,
		CreatedAt:        time.Unix(c.CreatedAt, 0),
	}
}
