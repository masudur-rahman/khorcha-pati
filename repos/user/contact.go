package user

import (
	"context"
	"fmt"
	"time"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"

	"github.com/masudur-rahman/styx"
	isql "github.com/masudur-rahman/styx/sql"
)

type SQLContactRepository struct {
	db     isql.Engine
	logger logr.Logger
}

func NewSQLContactRepository(db isql.Engine, logger logr.Logger) *SQLContactRepository {
	return &SQLContactRepository{
		db:     db.Table(models.Contacts{}.TableName()),
		logger: logger,
	}
}

func (u *SQLContactRepository) WithUnitOfWork(uow styx.UnitOfWork) repos.ContactRepository {
	return &SQLContactRepository{
		db:     uow.SQL.Table(models.Contacts{}.TableName()),
		logger: u.logger,
	}
}

func (u *SQLContactRepository) GetContactByID(id int64) (*models.Contacts, error) {
	u.logger.Infow("get contact by id", "id", id)
	ctx := context.Background()
	var c models.Contacts
	found, err := u.db.ID(id).FindOne(ctx, &c)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.ErrContactNotFound{}
	}
	return &c, nil
}

func (u *SQLContactRepository) GetContactByName(userID int64, name string) (*models.Contacts, error) {
	u.logger.Infow("get contact by name", "userID", userID, "name", name)
	if name == "" {
		return nil, models.ErrContactNotFound{UserID: userID}
	}
	ctx := context.Background()
	var c models.Contacts
	found, err := u.db.Where("LOWER(nick_name) = LOWER(?)", name).FindOne(ctx, &c, models.Contacts{UserID: userID})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.ErrContactNotFound{UserID: userID, NickName: name}
	}
	return &c, nil
}

func (u *SQLContactRepository) UpdateContactBalance(id int64, txnAmount float64) error {
	u.logger.Infow("updating contact balance", "id", id)
	ctx := context.Background()
	c, err := u.GetContactByID(id)
	if err != nil {
		return err
	}
	c.NetBalance += txnAmount
	c.LastTxnTimestamp = time.Now().Unix()
	return u.db.ID(c.ID).MustCols("net_balance", "last_txn_timestamp").UpdateOne(ctx, c)
}

func (u *SQLContactRepository) AddNewContact(contact *models.Contacts) error {
	if contact.UserID == 0 {
		return fmt.Errorf("user-id can't be empty")
	}
	ctx := context.Background()
	_, err := u.GetContactByName(contact.UserID, contact.NickName)
	if err == nil {
		return models.ErrContactAlreadyExist{UserID: contact.UserID, NickName: contact.NickName}
	} else if !models.IsErrNotFound(err) {
		return err
	}
	_, err = u.db.MustCols("net_balance", "last_txn_timestamp", "created_at").InsertOne(ctx, contact)
	return err
}

func (u *SQLContactRepository) ListContacts(userID int64) ([]models.Contacts, error) {
	u.logger.Infow("list contacts", "userID", userID)
	ctx := context.Background()
	contacts := make([]models.Contacts, 0)
	err := u.db.FindMany(ctx, &contacts, models.Contacts{UserID: userID})
	return contacts, err
}

func (u *SQLContactRepository) DeleteContact(id int64) error {
	u.logger.Infow("deleting contact", "id", id)
	ctx := context.Background()
	return u.db.DeleteOne(ctx, models.Contacts{ID: id})
}

func (u *SQLContactRepository) UpdateContact(contact *models.Contacts) error {
	u.logger.Infow("updating contact name and nickname", "id", contact.ID)
	ctx := context.Background()
	return u.db.ID(contact.ID).MustCols("nick_name", "full_name", "email", "contact_info").UpdateOne(ctx, contact)
}
