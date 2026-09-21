package crud

import (
	Config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
)

type SSHKeyCRUD struct {
	Repo *repository.SSHKeyRepository
}

func NewSSHKeyCRUD(repo *repository.SSHKeyRepository) *SSHKeyCRUD {
	if repo == nil {
		repo = repository.NewSSHKeyRepository(Config.DB)
	}
	return &SSHKeyCRUD{Repo: repo}
}

func (c *SSHKeyCRUD) CreateSSHKey(key *models.SSHKey) error {
	return c.Repo.CreateSSHKey(key)
}

func (c *SSHKeyCRUD) GetSSHKeys(page uint, limit uint) ([]models.SSHKey, uint, error) {
	return c.Repo.GetSSHKeys(page, limit)
}

func (c *SSHKeyCRUD) GetAllSSHKeys() ([]models.SSHKey, error) {
	return c.Repo.GetAllSSHKeys()
}

func (c *SSHKeyCRUD) GetSSHKeyByID(id uint) (*models.SSHKey, error) {
	return c.Repo.GetSSHKeyByID(id)
}

func (c *SSHKeyCRUD) GetSSHKeyByFingerprint(fingerprint string) (*models.SSHKey, error) {
	return c.Repo.GetSSHKeyByFingerprint(fingerprint)
}

func (c *SSHKeyCRUD) UpdateSSHKey(key *models.SSHKey, id uint) error {
	return c.Repo.UpdateSSHKey(key, id)
}

func (c *SSHKeyCRUD) DeleteSSHKey(id uint) error {
	return c.Repo.DeleteSSHKey(id)
}
