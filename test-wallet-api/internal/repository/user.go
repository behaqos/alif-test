package repository

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type userRepo struct {
	r *redis.Client
}

type UserRepo interface {
	GetUserByLogin(context.Context, string) (bool, error)
}

func NewUserRepo(r *redis.Client) *userRepo {
	return &userRepo{r}
}

func (u *userRepo) GetUserByLogin(ctx context.Context, login string) (bool, error) {
	r := u.r.Get(ctx, login)
	if r.Err() != nil {
		return false, r.Err()
	}
	return true, nil
}
