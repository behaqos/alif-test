package repository

import (
	"github.com/sheryorov/test-wallet-api-auth/internal/entity"
	"gorm.io/gorm"
)

type userRepo struct {
	*gorm.DB
}

func NewUserRepo(conn *gorm.DB) *userRepo {
	return &userRepo{conn}
}

func (u *userRepo) CheckUser(login string) (*entity.User, error) {
	user := entity.User{}
	res := u.Where(&entity.User{Login: login}).Find(&user)
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}
