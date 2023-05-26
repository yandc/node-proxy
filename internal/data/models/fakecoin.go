package models

type FakeCoinWhiteList struct {
	BaseModel
	Symbol  string `gorm:"index"`
	Name    string
	Chain   string `gorm:"index:,unique,composite:unique_chain_address"`
	Address string `gorm:"index:,unique,composite:unique_chain_address"`
}
