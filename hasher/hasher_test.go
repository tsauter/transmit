package hasher

import (
	"testing"
)

func Test_Interface(t *testing.T) {
	// make sure all hash implementations satisify the
	// Hasher interface
	var hash Hasher
	hash = &MD5Hasher{}
	hash = &SHA1Hasher{}
	hash = &SHA256Hasher{}

	if hash == nil {
		t.Errorf("hash is nil")
	}
}
