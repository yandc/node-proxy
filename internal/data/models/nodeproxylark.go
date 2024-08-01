package models

type NodeProxyLark struct {
	BaseModel
	Chain   string `gorm:"index:,unique,composite:unique_chain_address"`
	Address string `gorm:"index:,unique,composite:unique_chain_address"`
	ErrMsg  string
	ErrType string //token/nft
	TokenId string
}

const (
	ERR_TYPE_TOKEN = "token"
	ERR_TYPE_NFT   = "nft"
)
