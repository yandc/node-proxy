package models

type TokenInfo struct {
	BaseModel
	Chain    string `gorm:"index:,unique,composite:unique_chain_address"`
	Name     string
	Address  string `gorm:"index:,unique,composite:unique_chain_address"`
	Decimals uint32
	Symbol   string
	LogoURI  string
}
