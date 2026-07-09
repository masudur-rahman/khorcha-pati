package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/pkg/validator"
	"github.com/masudur-rahman/khorcha-pati/repos"
	"github.com/masudur-rahman/khorcha-pati/services"

	"github.com/masudur-rahman/styx"
)

type contactService struct {
	uow         styx.UnitOfWork
	contactRepo repos.ContactRepository
	txnRepo     repos.TransactionRepository
}

var _ services.ContactService = &contactService{}

func NewContactService(uow styx.UnitOfWork, contactRepo repos.ContactRepository, txnRepo repos.TransactionRepository) *contactService {
	return &contactService{
		uow:         uow,
		contactRepo: contactRepo,
		txnRepo:     txnRepo,
	}
}

func (cs *contactService) GetContactByID(id int64) (*models.Contacts, error) {
	return cs.contactRepo.GetContactByID(id)
}

func (cs *contactService) GetContactByName(userID int64, name string) (*models.Contacts, error) {
	return cs.contactRepo.GetContactByName(userID, name)
}

func (cs *contactService) ListContacts(userID int64) ([]models.Contacts, error) {
	return cs.contactRepo.ListContacts(userID)
}

func (cs *contactService) CreateContact(contact *models.Contacts) error {
	if contact.UserID == 0 {
		return fmt.Errorf("user-id can't be empty")
	}

	if !validator.IsValidShortName(contact.NickName) {
		return models.StatusError{
			Status:  400,
			Message: "contact nickname cannot have spaces or special characters",
		}
	}

	if contact.FullName != "" && !validator.IsValidDisplayName(contact.FullName) {
		return models.StatusError{
			Status:  400,
			Message: "contact full name cannot have leading/trailing spaces or special characters",
		}
	}

	contacts, err := cs.contactRepo.ListContacts(contact.UserID)
	if err != nil {
		return err
	}
	for _, existing := range contacts {
		if strings.EqualFold(existing.NickName, contact.NickName) {
			return models.StatusError{
				Status:  409,
				Message: fmt.Sprintf("contact already exists with nickname: %s", contact.NickName),
			}
		}
		if contact.FullName != "" && strings.EqualFold(existing.FullName, contact.FullName) {
			return models.StatusError{
				Status:  409,
				Message: fmt.Sprintf("contact already exists with full name: %s", contact.FullName),
			}
		}
	}

	contact.CreatedAt = time.Now().Unix()
	return cs.contactRepo.AddNewContact(contact)
}

func (cs *contactService) UpdateContact(userID, id int64, nickName, fullName, email string) (err error) {
	if nickName == "" {
		return models.StatusError{
			Status:  400,
			Message: "contact nickname cannot be empty",
		}
	}

	if !validator.IsValidShortName(nickName) {
		return models.StatusError{
			Status:  400,
			Message: "contact nickname cannot have spaces or special characters",
		}
	}

	if fullName != "" && !validator.IsValidDisplayName(fullName) {
		return models.StatusError{
			Status:  400,
			Message: "contact full name cannot have leading/trailing spaces or special characters",
		}
	}

	contact, err := cs.contactRepo.GetContactByID(id)
	if err != nil {
		return err
	}

	if contact.UserID != userID {
		return models.StatusError{
			Status:  403,
			Message: "contact does not belong to user",
		}
	}

	oldNickName := contact.NickName

	contacts, err := cs.contactRepo.ListContacts(userID)
	if err != nil {
		return err
	}
	for _, existing := range contacts {
		if existing.ID != id {
			if strings.EqualFold(existing.NickName, nickName) {
				return models.StatusError{
					Status:  409,
					Message: fmt.Sprintf("contact already exists with nickname: %s", nickName),
				}
			}
			if fullName != "" && strings.EqualFold(existing.FullName, fullName) {
				return models.StatusError{
					Status:  409,
					Message: fmt.Sprintf("contact already exists with full name: %s", fullName),
				}
			}
		}
	}

	contact.NickName = nickName
	contact.FullName = fullName
	contact.Email = email

	uow, err := cs.uow.Begin(context.Background())
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

	if err = cs.contactRepo.WithUnitOfWork(uow).UpdateContact(contact); err != nil {
		return err
	}

	if oldNickName != nickName {
		if err = cs.txnRepo.WithUnitOfWork(uow).UpdateTransactionsContact(userID, oldNickName, nickName); err != nil {
			return err
		}
	}

	return nil
}

func (cs *contactService) UpdateContactBalance(id int64, amount float64) error {
	return cs.contactRepo.UpdateContactBalance(id, amount)
}

func (cs *contactService) DeleteContact(id int64) error {
	contact, err := cs.contactRepo.GetContactByID(id)
	if err != nil {
		return err
	}

	// Check if transactions exist referencing this contact
	txns, err := cs.txnRepo.ListTransactions(models.Transaction{UserID: contact.UserID, ContactName: contact.NickName})
	if err != nil {
		return err
	}
	if len(txns) > 0 {
		return models.StatusError{
			Status:  400,
			Message: fmt.Sprintf("cannot delete contact '%s': it has active transactions", contact.NickName),
		}
	}

	return cs.contactRepo.DeleteContact(id)
}
