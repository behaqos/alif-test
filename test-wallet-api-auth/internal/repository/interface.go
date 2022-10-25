package repository

import (
	"context"

	"github.com/sheryorov/test-wallet-api-auth/internal/entity"
)

type Tokenizer interface {
	RegisterToken(context.Context, string, string) error
	RevokeToken(context.Context, string) error
	SignToken(string) (string, error)
	GetToken(context.Context, string) (string, error)
	Parse(string) error
}

type UserRepo interface {
	CheckUser(string) (*entity.User, error)
}
