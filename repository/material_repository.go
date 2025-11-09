package repository

import (
	"backend-avanzada/models"
	"errors"

	"gorm.io/gorm"
)

type MaterialRepository struct{ db *gorm.DB }

func NewMaterialRepository(db *gorm.DB) *MaterialRepository { return &MaterialRepository{db} }

func (r *MaterialRepository) FindAll() ([]*models.Material, error) {
	var list []*models.Material
	return list, r.db.Find(&list).Error
}
func (r *MaterialRepository) FindById(id int) (*models.Material, error) {
	var m models.Material
	err := r.db.Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}
func (r *MaterialRepository) Save(m *models.Material) (*models.Material, error) {
	return m, r.db.Save(m).Error
}
func (r *MaterialRepository) Delete(m *models.Material) error {
	return r.db.Delete(m).Error
}
