package user

import "task_mng/domain/user/entity"

type Repository interface {
	Create(e *entity.User) error
	FindByEmail(email string) (entity.User, error)
	FindByUsername(username string) (entity.User, error)
	FindByID(id uint) (entity.User, error)
	FindByIDs(ids []uint) ([]entity.User, error)
	FindAll(page, limit int) ([]entity.User, int64, error)
	Update(e entity.User) error
	Delete(id uint) error
}
