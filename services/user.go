package services

import (
	"github.com/masudur-rahman/expense-tracker-bot/models"
)

type ContactService interface {
	GetContactByID(id int64) (*models.Contacts, error)
	GetContactByName(userID int64, name string) (*models.Contacts, error)
	ListContacts(userID int64) ([]models.Contacts, error)
	CreateContact(contact *models.Contacts) error
	UpdateContactBalance(id int64, amount float64) error
	DeleteContact(id int64) error
}

type ProfileService interface {
	GetUserByID(id int64) (*models.Profile, error)
	GetUserByTelegramID(id int64) (*models.Profile, error)
	GetUserByUsername(username string) (*models.Profile, error)
	GetUserByIdentifier(identifier string) (*models.Profile, error)
	ListUsers() ([]models.Profile, error)
	SignUp(user *models.Profile) error
	UpdateUser(id int64, user *models.Profile) error
	UpdateMobileNumber(userID int64, mobile string) error
	DeleteUser(id int64) error
}
