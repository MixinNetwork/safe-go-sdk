package safe

import (
	"context"
	"testing"

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
}
