package repos

import (
	"github.com/masudur-rahman/khorcha-pati/models"

	"github.com/masudur-rahman/styx"
)

type ContactRepository interface {
	WithUnitOfWork(uow styx.UnitOfWork) ContactRepository
	GetContactByID(id int64) (*models.Contacts, error)
	GetContactByName(userID int64, name string) (*models.Contacts, error)
	ListContacts(userID int64) ([]models.Contacts, error)
	AddNewContact(contact *models.Contacts) error
	UpdateContact(contact *models.Contacts) error
	UpdateContactBalance(id int64, amount float64) error
	DeleteContact(id int64) error
}

type UserRepository interface {
	GetUserByID(id int64) (*models.Profile, error)
	GetUser(filter models.Profile) (*models.Profile, error)
	ListUsers() ([]models.Profile, error)
	AddNewUser(user *models.Profile) error
	UpdateUser(id int64, user *models.Profile) error
	SetActive(id int64, active bool) error
	DeleteUser(id int64) error
}
