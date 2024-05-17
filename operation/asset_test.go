package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsset(t *testing.T) {
	assert := assert.New(t)

	SetFactoryContractAddress(TestFactoryContractAddress)
	id, err := GetSafeBTCAssetId("0x9d04735aaEB73535672200950fA77C2dFC86eB21", "b7938396-3f94-4e0a-9179-d3440718156f", "03cf716e609663da611d84f1133c8440b9e2347960443eb3543fcc16faff43bd7a", "", "")
	assert.Nil(err)
	assert.Equal("e6d5bd3d-ebfa-3dd4-9017-ab0ae908db8f", id)
}
