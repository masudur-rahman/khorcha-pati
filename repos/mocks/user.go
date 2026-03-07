package mocks

import (
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/repos"

	"github.com/masudur-rahman/styx"

	"github.com/stretchr/testify/mock"
)

// UserRepo is a mock for repos.UserRepository.
type UserRepo struct {
	mock.Mock
}

var _ repos.UserRepository = &UserRepo{}

func (m *UserRepo) GetUserByID(id int64) (*models.Profile, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Profile), args.Error(1)
}

func (m *UserRepo) GetUser(filter models.Profile) (*models.Profile, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Profile), args.Error(1)
}

func (m *UserRepo) ListUsers() ([]models.Profile, error) {
	args := m.Called()
	return args.Get(0).([]models.Profile), args.Error(1)
}

func (m *UserRepo) AddNewUser(user *models.Profile) error {
	return m.Called(user).Error(0)
}

func (m *UserRepo) UpdateUser(id int64, user *models.Profile) error {
	return m.Called(id, user).Error(0)
}

func (m *UserRepo) DeleteUser(id int64) error {
	return m.Called(id).Error(0)
}

// ContactRepo is a mock for repos.ContactRepository.
type ContactRepo struct {
	mock.Mock
}

var _ repos.ContactRepository = &ContactRepo{}

func (m *ContactRepo) WithUnitOfWork(_ styx.UnitOfWork) repos.ContactRepository {
	return m
}

func (m *ContactRepo) GetContactByID(id int64) (*models.Contacts, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contacts), args.Error(1)
}

func (m *ContactRepo) GetContactByName(userID int64, name string) (*models.Contacts, error) {
	args := m.Called(userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contacts), args.Error(1)
}

func (m *ContactRepo) ListContacts(userID int64) ([]models.Contacts, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Contacts), args.Error(1)
}

func (m *ContactRepo) AddNewContact(contact *models.Contacts) error {
	return m.Called(contact).Error(0)
}

func (m *ContactRepo) UpdateContactBalance(id int64, amount float64) error {
	return m.Called(id, amount).Error(0)
}

func (m *ContactRepo) DeleteContact(id int64) error {
	return m.Called(id).Error(0)
}

// EventRepo is a mock for repos.EventRepository.
type EventRepo struct {
	mock.Mock
}

var _ repos.EventRepository = &EventRepo{}

func (m *EventRepo) AddEvent(event string) error {
	return m.Called(event).Error(0)
}

func (m *EventRepo) ListEvents() ([]models.Event, error) {
	args := m.Called()
	return args.Get(0).([]models.Event), args.Error(1)
}
