package opensea

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetOpenSeaNFTAsset(t *testing.T) {
	model, err := GetOpenSeaNFTAsset("Arbitrum", "0x43111161dc2eb245a0f51bb79310c1e80d0129b4", "19486")
	if err != nil {
		fmt.Println("error=", err)
	}
	b, _ := json.Marshal(model)
	fmt.Println("result=", string(b))
}
