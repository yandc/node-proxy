package aptosNFT

import (
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"testing"
)

func TestGetNFTAsset(t *testing.T) {
	nftInfo, err := GetAptosNFTAsset("Aptos", "0xf932dcb9835e681b21d2f411ef99f4f5e577e6ac299eebee2272a39fb348f702::Aptos Monkeys::0", "AptosMonkeys #3918")
	if err != nil {
		fmt.Println("error=", err)
	}
	fmt.Println("result=", nftInfo)
}

func TestGetNFTDataByTokenId(t *testing.T) {
	slug := getTopazNFTData("0xf932dcb9835e681b21d2f411ef99f4f5e577e6ac299eebee2272a39fb348f702::Aptos Monkeys::AptosMonkeys #5866::0")
	fmt.Println(slug)
}

func TestGetAptosCollection(t *testing.T) {
	result, err := nft.DoWebRequest("https://opensea.io/__api/graphql/")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("result==", result)
}
