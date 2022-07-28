package models

import (
	"gorm.io/gorm"
)

type CoinMarketCap struct {
	gorm.Model
	Id          int `gorm:"primaryKey"`
	Symbol      string
	Name        string
	Platform    string
	Category    string
	Twitter     string
	Logo        string
	WebSite     string
	Description string
}
