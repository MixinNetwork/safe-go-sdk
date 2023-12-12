package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountRPC(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	SetBaseUri(TestHost)
	account, err := ReadAccount(ctx, "59aabf15-7036-4ce2-9471-98f9aef147fc")
	assert.Nil(err)
	assert.Equal("59aabf15-7036-4ce2-9471-98f9aef147fc", account.ID)
	assert.Len(account.Outputs, 0)
	assert.Equal(int64(1), account.Chain)
}
