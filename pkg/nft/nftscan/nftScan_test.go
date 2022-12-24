package nftscan

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetNFTScanNFTAsset(t *testing.T) {
	model, err := GetNFTScanNFTAsset("BSC", "0x5cfbd085bdb4eb5097c829aadda69cad665b8031", "1")
	if err != nil {
		fmt.Println("error==", err)
	}
	b, _ := json.Marshal(model)
	fmt.Println("result=", string(b))
}

func TestGetNFTScanCollection(t *testing.T) {
	model, err := GetNFTScanCollection("BSC", "0x5cfbd085bdb4eb5097c829aadda69cad665b8031")
	if err != nil {
		fmt.Println("error==", err)
	}
	b, _ := json.Marshal(model)
	fmt.Println("result=", string(b))
}
