package list

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNFTListModel(t *testing.T) {
	re, err := GetNFTListModel("Ethereum", "0x8cc6517e45db7a0803fef220d9b577326a12033f", "22609")
	assert.NoError(t, err)
	b, _ := json.Marshal(re)
	fmt.Println("result=", string(b))
}
