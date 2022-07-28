package models

import "gorm.io/gorm"

type TokenList struct {
	gorm.Model
	Id          int
	CgId        string
	CmcId       int
	Symbol      string
	Name        string
	Chain       string `gorm:"primaryKey;index;index:idx_chain_address"`
	Address     string `gorm:"primaryKey;index:idx_chain_address"`
	Description string
	Logo        string
	LogoURI     string
	WebSite     string
	Decimals    int
}
