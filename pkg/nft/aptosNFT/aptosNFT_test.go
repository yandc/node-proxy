package aptosNFT

import (
	"fmt"
	"testing"
)

func TestGetNFTAsset(t *testing.T) {
	nftInfo, err := GetAptosNFTAsset("Aptos", "", "0x834048d84a968bd8aa1af5895903d8bd11168cc0cb7c9ce35eb549f6f4437bd6::HalloweenBoi::HalloweenBoi #385::0")
	if err != nil {
		fmt.Println("error=", err)
	}
	fmt.Println("result=", nftInfo)
}

func TestGetNFTDataByTokenId(t *testing.T) {
	slug := getTopazNFTData("0xf932dcb9835e681b21d2f411ef99f4f5e577e6ac299eebee2272a39fb348f702::Aptos Monkeys::AptosMonkeys #5866::0")
	fmt.Println(slug)
}
