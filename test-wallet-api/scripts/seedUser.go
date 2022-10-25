package scripts

import (
	"fmt"

	"github.com/sheryorov/test-wallet-api/internal/entity"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUser(db *gorm.DB) {
	wallets := []entity.Wallet{
		{
			Login: "wallet1",
			Sum:   10000,
		},
		{
			Login: "wallet2",
			Sum:   10500,
		},
		{
			Login: "wallet3",
			Sum:   9000,
		},
	}

	for i := range wallets {
		db.Create(&wallets[i])
	}
	users := map[string]string{
		"user1": "pwd1",
		"user2": "pwd2",
		"user3": "pwd3",
	}
	i := 0
	for user, password := range users {
		passwordBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		u := entity.User{
			Login:    user,
			Password: string(passwordBytes),
			WalletID: wallets[i].ID,
		}
		db.Create(&u)
		fmt.Println(u)
	}
}
