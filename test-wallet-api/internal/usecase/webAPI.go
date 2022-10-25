package usecase

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sheryorov/test-wallet-api/internal/repository"
	"github.com/sheryorov/test-wallet-api/pkg/auth"
	"go.uber.org/zap"
)

var ctx = context.Background()

type UserReq struct {
	Src         *uint    `json:",omitempty"`
	Dest        *uint    `json:",omitempty"`
	Sum         *float64 `json:",omitempty"`
	WalletLogin *string  `json:"wallet_login,omitempty"`
	WalletID    *uint    `json:"wallet_id,omitempty"`
	Login       *string  `json:",omitempty"`
	Password    *string  `json:",omitempty"`
}

type webApiHandler struct {
	walletRepo         repository.WalletRepo
	userRepo           repository.UserRepo
	transactionUsecase TransactionUsecase
	authClient         auth.AuthClient
	logger             *zap.Logger
}

func NewWebApiHandler(walletRepo repository.WalletRepo, tr TransactionUsecase, logger *zap.Logger, authClient auth.AuthClient, userRepo repository.UserRepo) *webApiHandler {
	return &webApiHandler{walletRepo: walletRepo, transactionUsecase: tr, logger: logger, authClient: authClient, userRepo: userRepo}
}

type WebApiHandler interface {
	CheckWallet(*gin.Context)
	Charge(*gin.Context)
	GetWalletHistory(*gin.Context)
	GetWalletBalance(*gin.Context)
	Login(*gin.Context)
}

func (w *webApiHandler) CheckWallet(c *gin.Context) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	reqBody := buf.Bytes()
	req := UserReq{}

	if err := json.Unmarshal(reqBody, &req); err != nil {
		fmt.Println(string(reqBody))
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	if _, err := w.walletRepo.GetWalletByLogin(*req.WalletLogin); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("wallet %s not found", *req.WalletLogin)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("wallet %s  exists", *req.WalletLogin)})
}

func (w *webApiHandler) Charge(c *gin.Context) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	reqBody := buf.Bytes()
	req := UserReq{}

	if err := json.Unmarshal(reqBody, &req); err != nil {
		fmt.Println(string(reqBody))
		w.logger.Sugar().Errorf("unmarshall error:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := w.transactionUsecase.ChargeWallet(*req.Src, *req.Dest, *req.Sum); err != nil {
		w.logger.Sugar().Errorf("charge error error:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment proccesed"})
}

func (w *webApiHandler) GetWalletHistory(c *gin.Context) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	reqBody := buf.Bytes()
	req := UserReq{}
	if err := json.Unmarshal(reqBody, &req); err != nil {
		fmt.Println(string(reqBody))
		w.logger.Sugar().Errorf("unmarshall error:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	walletHistory, err := w.walletRepo.GetWalletHistoryID(*req.WalletID)
	if err != nil {
		w.logger.Sugar().Errorf("unmarshall error:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": walletHistory})
}

func (w *webApiHandler) GetWalletBalance(c *gin.Context) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	reqBody := buf.Bytes()
	req := UserReq{}

	if err := json.Unmarshal(reqBody, &req); err != nil {
		fmt.Println(string(reqBody))
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	wallet, err := w.walletRepo.GetWalletByLogin(*req.WalletLogin)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("wallet %s not found", *req.WalletLogin)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("wallet %s  exists", *req.WalletLogin), "balance": wallet.Sum})
}

func (w *webApiHandler) Login(c *gin.Context) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	reqBody := buf.Bytes()
	req := UserReq{}

	if err := json.Unmarshal(reqBody, &req); err != nil {
		w.logger.Sugar().Errorf("unmarshall error:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := w.authClient.Authorize(ctx, &auth.UserRequest{Login: *req.Login, Password: *req.Password})
	if err != nil {
		w.logger.Sugar().Errorf("error on authorize:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": r})
}

func (w *webApiHandler) HeaderCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// middleware
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		digest := c.GetHeader("X-Digest")
		userID := c.GetHeader("X-UserId")
		b := sha1.New()
		b.Write(bodyBytes)
		sha := base64.URLEncoding.EncodeToString(b.Sum(nil))
		if sha != digest {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if valid, err := w.userRepo.GetUserByLogin(ctx, userID); !valid {
			w.logger.Sugar().Errorf("error on getting user:%v", err)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
