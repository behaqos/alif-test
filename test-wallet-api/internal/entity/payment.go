package entity

import "gorm.io/gorm"

type Payment struct {
	gorm.Model
	Src         uint
	Dst         uint
	Sum         float64
	Status      int
	Description string
}
