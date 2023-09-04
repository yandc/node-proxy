package models

import (
	"gorm.io/gorm"
)

type CoinGecko struct {
	gorm.Model
	Id            string `gorm:"primaryKey"`
	Symbol        string
	Name          string
	Platform      string
	Image         string
	Description   string
	Homepage      string
	CoinGeckoRank uint16
}

type CoinGeckoList struct {
	BaseModel
	CgId   string `gorm:"primaryKey"`
	Symbol string
	Name   string
}
