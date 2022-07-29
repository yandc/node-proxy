package models

type TokenList struct {
	BaseModel
	CgId        string
	CmcId       int
	Symbol      string
	Name        string
	Chain       string `gorm:"index:,unique,composite:unique_chain_address"`
	Address     string `gorm:"index:,unique,composite:unique_chain_address"`
	Description string
	Logo        string
	WebSite     string
	Decimals    int
	LogoURI     string
}
