package models

type NftCollection struct {
	BaseModel
	Chain       string
	Address     string
	Name        string
	Slug        string
	ImageURL    string
	Description string
	TokenType   string
}
