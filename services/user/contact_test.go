package user

import (
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos/mocks"

	"github.com/masudur-rahman/styx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testUserID int64 = 42

func TestCreateContact_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.ContactRepo{}
	txnRepo := &mocks.TransactionRepo{}
	svc := NewContactService(styx.UnitOfWork{}, repo, txnRepo)

	contact := &models.Contacts{
		UserID:   testUserID,
		NickName: "johndoe",
		FullName: "John Doe",
	}

	repo.On("ListContacts", testUserID).Return([]models.Contacts{}, nil)
	repo.On("AddNewContact", contact).Return(nil)

	err := svc.CreateContact(contact)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreateContact_duplicateError(t *testing.T) {
	t.Parallel()
	repo := &mocks.ContactRepo{}
	txnRepo := &mocks.TransactionRepo{}
	svc := NewContactService(styx.UnitOfWork{}, repo, txnRepo)

	existing := []models.Contacts{
		{UserID: testUserID, NickName: "johndoe", FullName: "John Doe"},
	}

	repo.On("ListContacts", testUserID).Return(existing, nil)

	contact := &models.Contacts{
		UserID:   testUserID,
		NickName: "JohnDoe", // duplicate nickname case-insensitive
		FullName: "Other Name",
	}

	err := svc.CreateContact(contact)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeleteContact_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.ContactRepo{}
	txnRepo := &mocks.TransactionRepo{}
	svc := NewContactService(styx.UnitOfWork{}, repo, txnRepo)

	contact := &models.Contacts{
		ID:       1,
		UserID:   testUserID,
		NickName: "johndoe",
	}

	repo.On("GetContactByID", int64(1)).Return(contact, nil)
	txnRepo.On("ListTransactions", mock.Anything).Return([]models.Transaction{}, nil)
	repo.On("DeleteContact", int64(1)).Return(nil)

	err := svc.DeleteContact(1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	txnRepo.AssertExpectations(t)
}

func TestDeleteContact_activeTransactionsError(t *testing.T) {
	t.Parallel()
	repo := &mocks.ContactRepo{}
	txnRepo := &mocks.TransactionRepo{}
	svc := NewContactService(styx.UnitOfWork{}, repo, txnRepo)

	contact := &models.Contacts{
		ID:       1,
		UserID:   testUserID,
		NickName: "johndoe",
	}

	repo.On("GetContactByID", int64(1)).Return(contact, nil)
	txnRepo.On("ListTransactions", mock.Anything).Return([]models.Transaction{
		{ID: 100},
	}, nil)

	err := svc.DeleteContact(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has active transactions")
}
