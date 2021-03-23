package short

import (
	"crypto/md5"
	"encoding/hex"
)

// Hasher wraps a method to "hashify" a string
type Hasher interface {
	Hash(text string) string
}

// MD5 is a hasher implemented using md5 message-digest algorithm
type MD5 struct{}

// NewMD5 returns a properly initialized MD5 hasher
func NewMD5() *MD5 {
	return &MD5{}
}

// Hash returns the digest produced from the input string
func (hasher *MD5) Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
