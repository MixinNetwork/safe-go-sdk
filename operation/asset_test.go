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

	id, err = GetSafeBTCAssetId("0x9d04735aaEB73535672200950fA77C2dFC86eB21", "c94ac88f-4671-3976-b60a-09064f1811e8", "039e2eca8938ba343daf7d32609c236c35f804b28be35a2ace527cd3f756455e70", "XIN", "Mixin")
	assert.Nil(err)
	assert.Equal("a43d9ab2-fee5-3f73-8175-58c3f0818aa3", id)
}
