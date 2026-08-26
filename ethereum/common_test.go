package ethereum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRecoveryRequestId(t *testing.T) {
	id := GetRecoveryRequestId(
		"0x346607eb15821A4E194628444F3705c26C8E6eBe",
		"0xA03A8590BB3A2cA5c747c8b99C63DA399424a055",
	)
	assert.Equal(t, "f4116e1d-6668-3593-b0b0-1644bb442d3e", id)
}
