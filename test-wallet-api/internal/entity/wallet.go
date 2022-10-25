package entity

import "gorm.io/gorm"

type Wallet struct {
	gorm.Model
	Login      string
	Sum        float64
	Identified bool
}

func (w *Wallet) BeforeUpdate(tx *gorm.DB) (err error) {
	if w.Sum < 0 {
		return tx.Rollback().Error
	}
	if w.Identified && w.Sum > 100000 {
		return tx.Rollback().Error
	}
	if !w.Identified && w.Sum > 10000 {
		return tx.Rollback().Error
	}
	return nil
}
