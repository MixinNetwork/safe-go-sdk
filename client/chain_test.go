package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainRPC(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	chains, err := ReadChains(ctx)
	assert.Nil(err)
	for i, c := range chains {
		if i == 0 {
			assert.Equal(int64(1), c.Chain)
		}
		if i == 1 {
			assert.Equal(int64(5), c.Chain)
		}
	}
}
