package common

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/gofrs/uuid/v5"
)

func DecodeHexOrPanic(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(s)
	}
	return b
}

func EncodeMixinExtra(appId, traceId, memo string) string {
	gid, err := uuid.FromString(appId)
	if err != nil {
		panic(err)
	}
	tid, err := uuid.FromString(traceId)
	if err != nil {
		panic(err)
	}
	var data []byte
	data = append(data, gid.Bytes()...)
	data = append(data, tid.Bytes()...)
	data = append(data, []byte(memo)...)
	return base64.RawURLEncoding.EncodeToString(data)
}
