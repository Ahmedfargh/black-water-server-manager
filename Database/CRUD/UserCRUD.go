package crud

import (
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
)

type UserCRUD struct {
	Repo *repository.UserRepository
}

func NewUserCRUD(repo *repository.UserRepository) *UserCRUD {
	return &UserCRUD{Repo: repo}
}

func (c *UserCRUD) CreateUser(user *models.User) error {
	return c.Repo.CreateUser(user)
}

func (c *UserCRUD) GetUserByID(id uint) (*models.User, error) {
	return c.Repo.GetUserByID(id)
}

func (c *UserCRUD) GetUsers() ([]models.User, error) {
	return c.Repo.GetUsers()
}

func (c *UserCRUD) GetPaginatedUsers(page, limit int) ([]models.User, int64, error) {
	return c.Repo.GetPaginatedUsers(page, limit)
}

func (c *UserCRUD) UpdateUser(user *models.User, id uint) error {
	return c.Repo.UpdateUser(user, id)
}

func (c *UserCRUD) DeleteUser(user *models.User) error {
	return c.Repo.DeleteUser(user)
}

func (c *UserCRUD) GetUserByUsername(username string) (*models.User, error) {
	return c.Repo.GetUserByUsername(username)
}

func (c *UserCRUD) GetUserByEmail(email string) (*models.User, error) {
	return c.Repo.GetUserByEmail(email)
}
