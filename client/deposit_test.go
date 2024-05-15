package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDepositRPC(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	deposits, err := ReadDeposits(ctx, 2, 0)
	assert.Nil(err)
	assert.True(len(deposits) > 0)
}
