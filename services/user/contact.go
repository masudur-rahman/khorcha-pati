package user

import (
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"
	"github.com/masudur-rahman/expense-tracker-bot/services"
)

type contactService struct {
	contactRepo repos.ContactRepository
}

var _ services.ContactService = &contactService{}

func NewContactService(contactRepo repos.ContactRepository) *contactService {
	return &contactService{contactRepo: contactRepo}
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
	return cs.contactRepo.AddNewContact(contact)
}

func (cs *contactService) UpdateContactBalance(id int64, amount float64) error {
	return cs.contactRepo.UpdateContactBalance(id, amount)
}

func (cs *contactService) DeleteContact(id int64) error {
	return cs.contactRepo.DeleteContact(id)
}
