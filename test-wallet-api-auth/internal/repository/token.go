package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
)

var (
	errTokenExpired = errors.New("token is expired")
)

type tokenizer struct {
	passPhrase  string
	expiration  time.Duration
	redisClient *redis.Client
}

func NewTokinizer(passPhrase string, dur time.Duration, r *redis.Client) *tokenizer {
	return &tokenizer{passPhrase: passPhrase, expiration: dur, redisClient: r}
}

func (t *tokenizer) RegisterToken(ctx context.Context, login string, token string) error {
	res := t.redisClient.Set(ctx, login, token, t.expiration)
	return res.Err()
}

func (t *tokenizer) RevokeToken(ctx context.Context, login string) error {
	res := t.redisClient.Del(ctx, login)
	return res.Err()
}

func (t *tokenizer) GetToken(ctx context.Context, login string) (string, error) {
	res := t.redisClient.Get(ctx, login)
	err := res.Err()
	if err != nil {
		return "", err
	}
	return res.String(), nil
}

func (t *tokenizer) SignToken(username string) (string, error) {
	exprDate := time.Now().Add(t.expiration).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":   exprDate,
		"login": username,
	})
	signedToken, err := token.SignedString([]byte(t.passPhrase))
	return signedToken, err
}

func (t *tokenizer) Parse(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(t.passPhrase), nil
	})
	if err != nil {
		return err
	}
	_, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return errTokenExpired
	}
	return nil
}
