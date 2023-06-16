package opensea

import (
	"encoding/json"
	"fmt"
	"testing"
)

type openSeaReq struct {
	Chain       string
	TokenAddres string
	TokenId     string
}

//"Arbitrum", "BSC", "Polygon", "Solana":
func TestGetOpenSeaNFTAsset(t *testing.T) {
	req := []openSeaReq{
		{"Polygon", "0x4d544035500D7aC1B42329c70eb58E77f8249f0F", "17032613055"},
		{"Solana", "EN8RnqfmvSJxCE2hPtVWBMfr5b8RfWmQRXGWoypBxFGn", "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},
		{"BSC", "0x858dd1107399804b4f39afaa5bd37a8ea8c8f188", "174"},
		{"Arbitrum", "0xfae39ec09730ca0f14262a636d2d7c5539353752", "30076"}}
	for _, token := range req {
		model, err := GetOpenSeaNFTAsset(token.Chain, token.TokenAddres, token.TokenId)
		if err != nil {
			fmt.Println("error=", err)
		}
		b, _ := json.Marshal(model)
		fmt.Println(token.Chain, ",result=", string(b))
	}

}
