package main

import (
	"fmt"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sheryorov/test-wallet-api-auth/config"
	"github.com/sheryorov/test-wallet-api-auth/internal/entity"
	"github.com/sheryorov/test-wallet-api-auth/internal/repository"
	"github.com/sheryorov/test-wallet-api-auth/internal/usecase"
	"github.com/sheryorov/test-wallet-api-auth/pkg/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	sugar := logger.Sugar()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.HTTP.Port))
	if err != nil {
		sugar.Fatalf("failed to listen: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Username: config.Redis.Username,
		Password: config.Redis.Password,
	})
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.PG.HOST, config.PG.Port, config.PG.Username, config.PG.Password, config.PG.DbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		sugar.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		sugar.Fatalf("cannot get db instance:%v", err)
	}
	// Close
	defer sqlDB.Close()
	db.AutoMigrate(&entity.User{})
	tokenizer := repository.NewTokinizer(config.Tokenikzer.PassPhrase, time.Duration(config.Tokenikzer.Expiration)*time.Hour, rdb)
	userRepo := repository.NewUserRepo(db)
	server := usecase.NewServerHandler(tokenizer, userRepo)
	s := grpc.NewServer()
	auth.RegisterAuthServer(s, server)
	if err := s.Serve(lis); err != nil {
		sugar.Fatalf("failed to start grpc service: %v", err)
	}
}
