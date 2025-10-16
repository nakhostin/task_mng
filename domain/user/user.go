package user

import (
	"task_mng/domain/user/entity"
	"task_mng/pkg/postgres"
)

type repository struct {
	db *postgres.Database
}

func New(db *postgres.Database) Repository {
	return &repository{db: db}
}

func (r *repository) Create(e *entity.User) error {
	return r.db.Create(e).Error
}

func (r *repository) FindByEmail(email string) (entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *repository) FindByUsername(username string) (entity.User, error) {
	var user entity.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return user, err
}

func (r *repository) FindByID(id uint) (entity.User, error) {
	var user entity.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return user, err
}

func (r *repository) FindByIDs(ids []uint) ([]entity.User, error) {
	var users []entity.User
	if len(ids) == 0 {
		return users, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *repository) FindAll(page, limit int) ([]entity.User, int64, error) {
	var users []entity.User
	var count int64

	offset := (page - 1) * limit

	err := r.db.Model(&entity.User{}).Count(&count).Error
	if err != nil {
		return users, count, err
	}

	err = r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, count, err
}

func (r *repository) Update(e entity.User) error {
	return r.db.Save(&e).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&entity.User{}, id).Error
}
