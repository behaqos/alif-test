package usecase

import (
	"context"
	"time"

	"github.com/sheryorov/test-wallet-api-auth/internal/repository"
	"github.com/sheryorov/test-wallet-api-auth/pkg/auth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	auth.UnimplementedAuthServer
	tokenizer repository.Tokenizer
	userRepo  repository.UserRepo
}

func NewServerHandler(t repository.Tokenizer, u repository.UserRepo) *server {
	return &server{tokenizer: t, userRepo: u}
}

func (s *server) Authorize(ctx context.Context, in *auth.UserRequest) (*auth.UserResponse, error) {
	user, err := s.userRepo.CheckUser(in.Login)
	if err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, err
	}
	token, err := s.tokenizer.SignToken(in.Login)
	if err != nil {
		return nil, err
	}
	if err := s.tokenizer.RegisterToken(context.Background(), in.Login, token); err != nil {
		return nil, err
	}

	return &auth.UserResponse{Token: token, ExpiredAt: time.Now().Local().String()}, err

}

func (s *server) CheckToken(ctx context.Context, in *auth.TokenRequest) (*emptypb.Empty, error) {
	token, err := s.tokenizer.GetToken(ctx, in.Login)
	if err != nil {
		return nil, err
	}
	if err = s.tokenizer.Parse(token); err != nil {
		return nil, err
	}
	return nil, nil
}
