package entity

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FullName string `gorm:"not null"`
	Username string `gorm:"not null;unique"`
	Email    string `gorm:"not null"`
	Password string `gorm:"not null"`
}

func NewUser(username, fullName, email, password string) (User, error) {
	return User{
			Username: username,
			FullName: fullName,
			Email:    email,
			Password: password,
		},
		nil
}

func (User) TableName() string {
	return "users"
}
