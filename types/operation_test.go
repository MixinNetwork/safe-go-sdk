package types

import (
	"encoding/base64"
	"log"
	"testing"
)

func TestOperation(t *testing.T) {
	data, _ := base64.RawURLEncoding.DecodeString("Z7kWVV0JRoai4mQZOI6W5HABIQIlrhoBUav8RGhatnHAOH9S8yHhQZJbUYi1UKpDMVDTdU6KxAUZqVgxNadf-XjM2y1_YmMxcXQ3YTdhc3B6dGo2YTI3Z25qcjdnMHd4OXE2dmE4NTltdjk1NGE2dDZ4ZnZsajh5eG1leHFjNHlwOTI")
	op, err := DecodeOperation(data)
	log.Println(err)
	log.Printf("%#v", op)
	log.Printf("%x", op.Extra)
}