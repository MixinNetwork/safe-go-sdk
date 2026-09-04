package operation

import (
	"bytes"
	"testing"

	"github.com/MixinNetwork/mixin/crypto"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestECDHEd25519(t *testing.T) {
	alice := crypto.NewKeyFromSeed(bytes.Repeat([]byte{1}, 64))
	bob := crypto.NewKeyFromSeed(bytes.Repeat([]byte{2}, 64))
	alicePublic := alice.Public()
	bobPublic := bob.Public()

	aliceSecret := ECDHEd25519(alice.String(), bobPublic.String())
	bobSecret := ECDHEd25519(bob.String(), alicePublic.String())
	assert.Equal(t, aliceSecret, bobSecret)
	assert.NotEqual(t, [32]byte{}, aliceSecret)

	assert.Panics(t, func() { ECDHEd25519("invalid", bobPublic.String()) })
	assert.Panics(t, func() { ECDHEd25519(alice.String(), "invalid") })
}

func TestAESEncryptDecrypt(t *testing.T) {
	secret := bytes.Repeat([]byte{0x42}, 32)
	id := uuid.Must(uuid.NewV4())
	plaintext := append(id.Bytes(), []byte("encrypted safe operation")...)

	ciphertext := AESEncrypt(secret, plaintext, id.String())
	assert.NotEqual(t, plaintext, ciphertext)
	assert.Greater(t, len(ciphertext), len(plaintext))
	assert.Equal(t, plaintext, AESDecrypt(secret, ciphertext))

	assert.PanicsWithValue(t, id.String(), func() {
		AESEncrypt(secret, []byte("short"), id.String())
	})
	otherID := uuid.Must(uuid.NewV4())
	assert.PanicsWithValue(t, otherID.String(), func() {
		AESEncrypt(secret, plaintext, otherID.String())
	})
	assert.Panics(t, func() {
		AESEncrypt([]byte("bad-key"), plaintext, id.String())
	})
	assert.Panics(t, func() {
		AESDecrypt([]byte("bad-key"), ciphertext)
	})

	tampered := append([]byte(nil), ciphertext...)
	tampered[len(tampered)-1] ^= 1
	assert.Panics(t, func() { AESDecrypt(secret, tampered) })
	assert.Panics(t, func() { AESDecrypt(secret, nil) })

	require.Equal(t, plaintext[:12], ciphertext[:12])
}
