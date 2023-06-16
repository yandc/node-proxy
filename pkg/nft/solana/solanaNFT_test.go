package solana

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSolanaNFTAsset(t *testing.T) {
	re, err := GetSolanaNFTAsset("Solana", "EN8RnqfmvSJxCE2hPtVWBMfr5b8RfWmQRXGWoypBxFGn", "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	assert.NoError(t, err)
	b, _ := json.Marshal(re)
	fmt.Println("result=", string(b))
}
