package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sheryorov/test-wallet-api/config"
	"github.com/sheryorov/test-wallet-api/internal/entity"
	"github.com/sheryorov/test-wallet-api/internal/repository"
	"github.com/sheryorov/test-wallet-api/internal/usecase"
	"github.com/sheryorov/test-wallet-api/pkg/auth"
	"github.com/sheryorov/test-wallet-api/scripts"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer logger.Sync() // flushes buffer, if any
	config, err := config.NewConfig()
	if err != nil {
		sugar.Fatalf("cannot read config:%v", err)
	}

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
	db.AutoMigrate(&entity.Payment{})
	db.AutoMigrate(&entity.User{})
	db.AutoMigrate(&entity.Wallet{})
	// Comment after first launch
	scripts.SeedUser(db)
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", config.HTTP.AuthHost, config.HTTP.AuthPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := auth.NewAuthClient(conn)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Username: config.Redis.Username,
		Password: config.Redis.Password,
	})
	walletRepo := repository.NewWalletRepo(db)
	transactionUsecase := usecase.NewTransactionUsecase(walletRepo, db)
	userRepo := repository.NewUserRepo(rdb)
	webApiHandler := usecase.NewWebApiHandler(walletRepo, transactionUsecase, logger, c, userRepo)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r := gin.New()
	r.POST("/login", webApiHandler.Login)
	v1 := r.Group("/v1")
	{
		v1.Use(webApiHandler.HeaderCheck())
		v1.POST("/checkwallet", webApiHandler.CheckWallet)
		v1.POST("/charge", webApiHandler.Charge)
		v1.POST("/gethistory", webApiHandler.GetWalletHistory)
		v1.POST("/getbalance", webApiHandler.GetWalletBalance)
	}

	srv := &http.Server{
		Addr:    config.HTTP.Port,
		Handler: r,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

}
