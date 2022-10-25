package repository

import (
	"github.com/sheryorov/test-wallet-api/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletRepo struct {
	*gorm.DB
}

type WalletRepo interface {
	GetWalletByID(uint, bool) (*entity.Wallet, error)
	GetWalletByLogin(string) (*entity.Wallet, error)
	GetWalletHistoryID(uint) ([]entity.Payment, error)
}

func NewWalletRepo(db *gorm.DB) *walletRepo {
	return &walletRepo{db}
}

func (w *walletRepo) GetWalletByID(walletID uint, locked bool) (*entity.Wallet, error) {
	wallet := entity.Wallet{}
	if locked {
		if err := w.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", walletID).Error; err != nil {
			return nil, err
		}
		return &wallet, nil
	}
	if err := w.Where("id = ?", walletID).Find(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (w *walletRepo) GetWalletHistoryID(walletID uint) ([]entity.Payment, error) {
	payments := []entity.Payment{}
	if err := w.Where("src = ? or dest = ?", walletID, walletID).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}
func (w *walletRepo) GetWalletByLogin(login string) (*entity.Wallet, error) {
	wallet := entity.Wallet{}
	if err := w.Where("login = ?", login).Find(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}
