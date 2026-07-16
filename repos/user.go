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
	FindUsers(filter models.Profile) ([]models.Profile, error)
	ListUsers() ([]models.Profile, error)
	AddNewUser(user *models.Profile) error
	UpdateUser(id int64, user *models.Profile) error
	SetActive(id int64, active bool) error
	DeleteUser(id int64) error
}

// FindUserByIdentifier resolves a username or phone number, tolerating a
// present-or-absent country code on the phone. Phone lookup queries the DB by
// the stored last-8-digits suffix, then verifies candidates against the full
// number so a same-suffix number from another country never matches.
func FindUserByIdentifier(repo UserRepository, identifier string) (*models.Profile, error) {
	user, err := repo.GetUser(models.Profile{Username: identifier})
	if err == nil {
		return user, nil
	}
	if !models.IsErrNotFound(err) {
		return nil, err
	}

	suffix := models.PhoneSuffix(identifier)
	if suffix == "" {
		return nil, err // no digits — preserve the username not-found error
	}
	candidates, ferr := repo.FindUsers(models.Profile{MobileSuffix: suffix})
	if ferr != nil {
		return nil, ferr
	}

	var matched *models.Profile
	for i := range candidates {
		if !models.PhoneNumbersMatch(candidates[i].MobileNumber, identifier) {
			continue
		}
		if matched != nil {
			// Ambiguous across accounts — require the country code instead of guessing.
			return nil, err
		}
		matched = &candidates[i]
	}
	if matched == nil {
		return nil, err
	}
	return matched, nil
}
