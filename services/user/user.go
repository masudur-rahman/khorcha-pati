package user

import (
	"time"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"
)

type userService struct {
	userRepo repos.UserRepository
}

func NewProfileService(userRepo repos.UserRepository) *userService {
	return &userService{userRepo: userRepo}
}

func (us *userService) GetUserByID(id int64) (*models.Profile, error) {
	return us.userRepo.GetUserByID(id)
}

func (us *userService) GetUserByTelegramID(id int64) (*models.Profile, error) {
	filter := models.Profile{TelegramID: id}
	return us.userRepo.GetUser(filter)
}

func (us *userService) GetUserByUsername(username string) (*models.Profile, error) {
	filter := models.Profile{Username: username}
	return us.userRepo.GetUser(filter)
}

// GetUserByIdentifier looks up a user by username first, then by mobile number.
func (us *userService) GetUserByIdentifier(identifier string) (*models.Profile, error) {
	user, err := us.userRepo.GetUser(models.Profile{Username: identifier})
	if err == nil {
		return user, nil
	}
	if !models.IsErrNotFound(err) {
		return nil, err
	}
	return us.userRepo.GetUser(models.Profile{MobileNumber: identifier})
}

func (us *userService) ListUsers() ([]models.Profile, error) {
	return us.userRepo.ListUsers()
}

func (us *userService) SignUp(user *models.Profile) error {
	user.CreatedAt = time.Now().Unix()
	user.IsActive = true
	return us.userRepo.AddNewUser(user)
}

// SetActive enables or disables a user. Disabled users are blocked from web auth and bot interaction.
func (us *userService) SetActive(userID int64, active bool) error {
	return us.userRepo.SetActive(userID, active)
}

func (us *userService) UpdateUser(id int64, user *models.Profile) error {
	return us.userRepo.UpdateUser(id, user)
}

// UpdateMobileNumber sets the mobile number for an existing user.
func (us *userService) UpdateMobileNumber(userID int64, mobile string) error {
	user, err := us.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}
	user.MobileNumber = mobile
	return us.userRepo.UpdateUser(userID, user)
}

func (us *userService) DeleteUser(id int64) error {
	return us.userRepo.DeleteUser(id)
}
