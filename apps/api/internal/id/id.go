package id

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

var encoding = base32.NewEncoding("0123456789ABCDEFGHJKMNPQRSTVWXYZ").WithPadding(base32.NoPadding)

func New() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return strings.ToUpper(encoding.EncodeToString(bytes)), nil
}
