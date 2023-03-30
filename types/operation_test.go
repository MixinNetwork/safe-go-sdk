package types

import (
	"encoding/base64"
	"log"
	"testing"
)

func TestOperation(t *testing.T) {
	data, _ := base64.RawURLEncoding.DecodeString("42IUB3KASRO2cbLJxNLdsm4BIQIlrhoBUav8RGhatnHAOH9S8yHhQZJbUYi1UKpDMVDTdRICAunluAf6i0VajfqxidKDEP8")
	op, err := DecodeOperation(data)
	log.Println(err)
	log.Printf("%#v", op)
	log.Printf("%x", op.Extra)
}