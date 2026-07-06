package user

import (
	"fmt"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos/mocks"

	"github.com/stretchr/testify/assert"
)

func TestGetUserByTelegramID_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	expected := &models.Profile{
		ID:         1,
		TelegramID: 123456,
		Username:   "testuser",
		FirstName:  "Test",
	}

	repo.On("GetUser", models.Profile{TelegramID: 123456}).Return(expected, nil)

	result, err := svc.GetUserByTelegramID(123456)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetUserByTelegramID_notFound(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	repo.On("GetUser", models.Profile{TelegramID: 999}).
		Return(nil, fmt.Errorf("not found"))

	result, err := svc.GetUserByTelegramID(999)

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestGetUserByUsername_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	expected := &models.Profile{
		ID:       1,
		Username: "testuser",
	}

	repo.On("GetUser", models.Profile{Username: "testuser"}).Return(expected, nil)

	result, err := svc.GetUserByUsername("testuser")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetUserByID_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	expected := &models.Profile{ID: 1, Username: "testuser"}

	repo.On("GetUserByID", int64(1)).Return(expected, nil)

	result, err := svc.GetUserByID(1)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSignUp_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	user := &models.Profile{
		TelegramID: 123456,
		Username:   "newuser",
		FirstName:  "New",
	}

	repo.On("AddNewUser", user).Return(nil)

	err := svc.SignUp(user)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestListUsers_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	expected := []models.Profile{
		{ID: 1, Username: "user1"},
		{ID: 2, Username: "user2"},
	}

	repo.On("ListUsers").Return(expected, nil)

	result, err := svc.ListUsers()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestDeleteUser_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	repo.On("DeleteUser", int64(1)).Return(nil)

	err := svc.DeleteUser(1)

	assert.NoError(t, err)
}

func TestUpdateUser_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	updated := &models.Profile{Username: "updated"}
	repo.On("UpdateUser", int64(1), updated).Return(nil)

	err := svc.UpdateUser(1, updated)

	assert.NoError(t, err)
}

func TestGetUserByIdentifier_foundByUsername(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	expected := &models.Profile{ID: 1, Username: "masud"}
	repo.On("GetUser", models.Profile{Username: "masud"}).Return(expected, nil)

	result, err := svc.GetUserByIdentifier("masud")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetUserByIdentifier_foundByMobile(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	expected := &models.Profile{ID: 1, MobileNumber: "+8801700000000"}
	repo.On("GetUser", models.Profile{Username: "+8801700000000"}).
		Return(nil, models.ErrUserNotFound{})
	repo.On("GetUser", models.Profile{MobileNumber: "+8801700000000"}).
		Return(expected, nil)

	result, err := svc.GetUserByIdentifier("+8801700000000")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetUserByIdentifier_notFound(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	repo.On("GetUser", models.Profile{Username: "nobody"}).
		Return(nil, models.ErrUserNotFound{})
	repo.On("GetUser", models.Profile{MobileNumber: "nobody"}).
		Return(nil, models.ErrUserNotFound{})

	result, err := svc.GetUserByIdentifier("nobody")

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestUpdateMobileNumber_success(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	existing := &models.Profile{ID: 1, Username: "masud", FirstName: "Masud"}
	repo.On("GetUserByID", int64(1)).Return(existing, nil)

	updated := &models.Profile{
		ID: 1, Username: "masud", FirstName: "Masud",
		MobileNumber: "+8801700000000",
	}
	repo.On("UpdateUser", int64(1), updated).Return(nil)

	err := svc.UpdateMobileNumber(1, "+8801700000000")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateMobileNumber_userNotFound(t *testing.T) {
	t.Parallel()
	repo := &mocks.UserRepo{}
	svc := NewProfileService(repo)

	repo.On("GetUserByID", int64(999)).
		Return(nil, models.ErrUserNotFound{})

	err := svc.UpdateMobileNumber(999, "+8801700000000")

	assert.Error(t, err)
}
