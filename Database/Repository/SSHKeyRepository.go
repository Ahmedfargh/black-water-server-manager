package Repository

import (
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type SSHKeyRepository struct {
	DB *gorm.DB
}

func NewSSHKeyRepository(db *gorm.DB) *SSHKeyRepository {
	return &SSHKeyRepository{DB: db}
}

func (r *SSHKeyRepository) CreateSSHKey(key *models.SSHKey) error {
	return r.DB.Create(key).Error
}

func (r *SSHKeyRepository) GetSSHKeys(page uint, limit uint) ([]models.SSHKey, uint, error) {
	var total int64
	var keys []models.SSHKey
	if err := r.DB.Model(&models.SSHKey{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := r.DB.Limit(int(limit)).Offset(int(offset)).Order("id desc").Find(&keys).Error
	if err != nil {
		return nil, 0, err
	}
	return keys, uint(total), nil
}

func (r *SSHKeyRepository) GetAllSSHKeys() ([]models.SSHKey, error) {
	var keys []models.SSHKey
	err := r.DB.Find(&keys).Error
	return keys, err
}

func (r *SSHKeyRepository) GetSSHKeyByID(id uint) (*models.SSHKey, error) {
	var key models.SSHKey
	err := r.DB.First(&key, id).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *SSHKeyRepository) GetSSHKeyByFingerprint(fingerprint string) (*models.SSHKey, error) {
	var key models.SSHKey
	err := r.DB.Where("fingerprint = ?", fingerprint).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *SSHKeyRepository) UpdateSSHKey(key *models.SSHKey, id uint) error {
	return r.DB.Model(&models.SSHKey{}).Where("id = ?", id).Updates(key).Error
}

func (r *SSHKeyRepository) DeleteSSHKey(id uint) error {
	return r.DB.Delete(&models.SSHKey{}, id).Error
}
