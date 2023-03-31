package client

import (
	"context"
	"log"
	"testing"

	"github.com/MixinNetwork/go-number"
	"github.com/stretchr/testify/assert"
)

func TestAccountRPC(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	account, err := ReadAccount(ctx, "cbceb4da-b0c4-4b81-9688-d40ee64eb868")
	assert.Nil(err)
	assert.Equal("cbceb4da-b0c4-4b81-9688-d40ee64eb868", account.ID)
	assert.Len(account.Accountant.Outputs, 1)
	assert.Len(account.Outputs, 1)
	assert.Equal(int64(1), account.Chain)

	account, err = ReadAccount(ctx, "ec6e0eb3-3f2d-4220-a6c2-612aaba51995")
	var feeTotal int64 = 0
	for _, output := range account.Accountant.Outputs {
		log.Println(output.Satoshi)
		feeTotal += output.Satoshi
	}
	log.Println(feeTotal)
	fee := number.NewDecimal(feeTotal, 8)
	log.Println(fee.Cmp(number.FromString("0.001")))
}
