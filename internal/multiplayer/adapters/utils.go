package adapters

import (
	"crypto/rand"
	"encoding/hex"
)

func generateQuestionID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
