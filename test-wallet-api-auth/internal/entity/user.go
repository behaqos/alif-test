package entity

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login     string
	Password  string
	PartnerID string
	WalletID  string
}
