package models

type BlockChain struct {
	BaseModel
	Name           string
	Chain          string
	Title          string `gorm:"index"`
	Logo           string
	ChainId        string `gorm:"index:,unique"`
	CurrencyName   string
	CurrencySymbol string
	Decimals       uint8
	Explorer       string
	ChainSlug      string
	IsTest         bool `gorm:"index"`
}

const (
	// 节点状态
	ChainNodeUrlStatusAvailable   = uint8(1)
	ChainNodeUrlStatusUnAvailable = uint8(2)

	// 节点来源
	ChainNodeUrlSourcePublic = uint8(1)
	ChainNodeUrlSourceCustom = uint8(2)
)

type ChainNodeUrl struct {
	BaseModel
	ChainId string `gorm:"index"`
	Url     string `gorm:"index:,unique"`
	Height  uint64 `gorm:"index"`
	Latency uint64 `gorm:"index"` //延迟(毫秒)
	Privacy string
	Status  uint8 `gorm:"index"` //1-可用 2-不可用
	InUsed  bool  `gorm:"index"`
	Source  uint8 `gorm:"index"` //1-公开推荐 2-用户自定义
}
