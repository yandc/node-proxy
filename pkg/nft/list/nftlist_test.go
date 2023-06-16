package list

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNFTListModel(t *testing.T) {
	re, err := GetNFTListModel("Solana", "2FMDW8tDSLtCgwfpvhoxSY2TbaAfAcZd1v9hny6JJu2Z", "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	assert.NoError(t, err)
	b, _ := json.Marshal(re)
	fmt.Println("result=", string(b))
}
