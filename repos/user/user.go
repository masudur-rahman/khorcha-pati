package user

import (
	"context"

	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"

	isql "github.com/masudur-rahman/styx/sql"
)

type SQLUserRepository struct {
	db     isql.Engine
	logger logr.Logger
}

func NewSQLUserRepository(db isql.Engine, logger logr.Logger) *SQLUserRepository {
	return &SQLUserRepository{
		db:     db.Table(models.Profile{}.TableName()),
		logger: logger,
	}
}

func (u *SQLUserRepository) GetUserByID(id int64) (*models.Profile, error) {
	u.logger.Infow("finding user by id", "id", id)
	ctx := context.Background()
	var user models.Profile
	found, err := u.db.ID(id).FindOne(ctx, &user)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.ErrUserNotFound{}
	}
	return &user, nil
}

func (u *SQLUserRepository) GetUser(filter models.Profile) (*models.Profile, error) {
	ctx := context.Background()
	var user models.Profile
	found, err := u.db.FindOne(ctx, &user, filter)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.ErrUserNotFound{}
	}
	return &user, nil
}

func (u *SQLUserRepository) GetUserByUsername(username string) (*models.Profile, error) {
	u.logger.Infow("finding user by name", "username", username)
	ctx := context.Background()
	filter := models.Profile{
		Username: username,
	}
	var user models.Profile
	found, err := u.db.FindOne(ctx, &user, filter)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, models.ErrUserNotFound{Username: username}
	}
	return &user, nil
}

func (u *SQLUserRepository) ListUsers() ([]models.Profile, error) {
	u.logger.Infow("listing users")
	ctx := context.Background()
	users := make([]models.Profile, 0)
	err := u.db.FindMany(ctx, &users)
	return users, err
}

func (u *SQLUserRepository) AddNewUser(user *models.Profile) error {
	ctx := context.Background()
	_, err := u.db.InsertOne(ctx, user)
	return err
}

func (u *SQLUserRepository) UpdateUser(id int64, us *models.Profile) error {
	ctx := context.Background()
	user, err := u.GetUserByID(id)
	if err != nil {
		return err
	}
	user.Username = us.Username
	user.FirstName = us.FirstName
	user.LastName = us.LastName
	user.MobileNumber = us.MobileNumber

	return u.db.ID(id).UpdateOne(ctx, user)
}

func (u *SQLUserRepository) DeleteUser(id int64) error {
	u.logger.Infow("deleting user", "id", id)
	ctx := context.Background()
	return u.db.DeleteOne(ctx, models.Profile{ID: id})
}
