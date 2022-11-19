package models

type NftList struct {
	BaseModel
	TokenId               string `gorm:"index:,unique,composite:unique_chain_address_id" json:"token_id"`
	TokenAddress          string `gorm:"index:,unique,composite:unique_chain_address_id" json:"token_address"`
	Name                  string `json:"name"`
	Symbol                string `json:"symbol"`
	Description           string `json:"description"`
	Chain                 string `gorm:"index:,unique,composite:unique_chain_address_id" json:"chain"`
	Network               string `json:"network"`
	TokenType             string `json:"token_type"`
	Rarity                string `json:"rarity"`             //稀有度
	Properties            string `json:"properties"`         //属性 traits
	ImageURL              string `json:"image_url"`          //保存在平台
	ImageOriginalURL      string `json:"image_original_url"` //原始图片
	CollectionName        string `json:"collection_name"`    //系列名称
	CollectionSlug        string `json:"collection_slug"`
	CollectionDescription string `json:"collection_description"`
	CollectionImageURL    string `json:"collection_image_url"`
	NftName               string `json:"nft_name"`
	AnimationURL          string `json:"animation_url"`
}
