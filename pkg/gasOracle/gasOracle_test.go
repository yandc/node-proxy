package gasOracle

import "testing"

func TestURLHost(t *testing.T) {
	host, err := urlHost("https://mempool.space/api/v1/fees/recommended")
	if err != nil {
		t.Fatal(err)
	}
	if host != "mempool.space" {
		t.Fail()
	}

	host, err = urlHost("https://api.blockcypher.com/v1/ltc/main")
	if err != nil {
		t.Fatal(err)
	}
	if host != "api.blockcypher.com" {
		t.Fail()
	}
}
