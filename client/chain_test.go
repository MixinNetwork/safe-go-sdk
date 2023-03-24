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
	for _, c := range chains {
		assert.Equal(int64(1), c.Chain)
	}
}
