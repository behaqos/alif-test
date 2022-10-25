package usecase

import (
	"net/http"

	"github.com/sheryorov/test-wallet-api/internal/entity"
	"github.com/sheryorov/test-wallet-api/internal/repository"
	"gorm.io/gorm"
)

type transactionUsecase struct {
	walletRepo repository.WalletRepo
	db         *gorm.DB
}

type TransactionUsecase interface {
	ChargeWallet(uint, uint, float64) error
}

func NewTransactionUsecase(walletRepo repository.WalletRepo, db *gorm.DB) *transactionUsecase {
	return &transactionUsecase{walletRepo: walletRepo, db: db}
}

func (t *transactionUsecase) ChargeWallet(src uint, dest uint, sum float64) error {
	return t.db.Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				payment := entity.Payment{
					Src:         src,
					Dst:         dest,
					Sum:         sum,
					Status:      http.StatusBadRequest,
					Description: "Error on handling payment",
				}
				t.db.Create(&payment)
			}
		}()
		walletSrc, err := t.walletRepo.GetWalletByID(src, true)
		if err != nil || (walletSrc.Sum-sum) < 0 {
			return tx.Rollback().Error
		}
		walletDest, err := t.walletRepo.GetWalletByID(dest, true)
		if err != nil {
			return tx.Rollback().Error
		}
		tx.Model(&walletSrc).UpdateColumn("sum", gorm.Expr("sum  - ?", sum))
		tx.Model(&walletDest).UpdateColumn("sum", gorm.Expr("sum  + ?", sum))
		payment := entity.Payment{
			Src:         src,
			Dst:         dest,
			Sum:         sum,
			Status:      http.StatusOK,
			Description: "Proccessed",
		}
		tx.Create(&payment)
		return nil
	})
}
