package operation

import "testing"

func TestSignTx(t *testing.T) {
	SignSafeTx("raw transaction hash data", "private key")
}
