package types

import (
	"encoding/base64"
	"log"
	"testing"
)

func TestOperation(t *testing.T) {
	data, _ := base64.RawURLEncoding.DecodeString("e-tTi-WnSaKbi7x1mXABRnABIQJ8yJT1uibdfYddBsRyztX2DMi2WxDKYAzSPeZcyHrMOk6R3K9n6II44LwBGWRTJ5SKYmMxcTN0NTJjd2E5MjYzMzAzeTZ1N245N2s0M21ja3d3NHFnNHRmNXE0aHphdGQ2dnR5d3phcnM1dXk4Zzc")
	op, err := DecodeOperation(data)
	log.Println(err)
	log.Printf("%#v", op)
	log.Printf("%x", op.Extra)
}