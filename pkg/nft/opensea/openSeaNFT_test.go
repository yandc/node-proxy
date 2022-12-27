package opensea

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetOpenSeaNFTAsset(t *testing.T) {
	model, err := GetOpenSeaNFTAsset("Polygon", "0x4d544035500D7aC1B42329c70eb58E77f8249f0F", "17032613055")
	if err != nil {
		fmt.Println("error=", err)
	}
	b, _ := json.Marshal(model)
	fmt.Println("result=", string(b))
}
