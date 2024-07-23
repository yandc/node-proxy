package models

type NodeProxyLark struct {
	BaseModel
	Chain   string `gorm:"index:,unique,composite:unique_chain_address"`
	Address string `gorm:"index:,unique,composite:unique_chain_address"`
	ErrMsg  string
}
