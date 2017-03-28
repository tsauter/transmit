package hasher

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

// Hasher structure to hold internal data.
type SHA1Hasher struct {
	TotalHash hash.Hash
}

// NewSHA1Hasher returns an initialized SHA1 hasher struct.
func NewSHA1Hasher() *SHA1Hasher {
	h := SHA1Hasher{}
	h.TotalHash = sha1.New()
	return &h
}

// GetName return the name or hash algorithm of this hasher implementation.
func (h *SHA1Hasher) GetName() string {
	return "SHA1"
}

// HashChunk takes a byte slice and create a checksum for these bytes.
func (h *SHA1Hasher) HashChunk(data []byte) string {
	// write to total hash
	h.TotalHash.Write(data)

	hash := sha1.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// GetFilehash returns the total checksum for all previously processed chunks.
// Usually this is the checksum of the file (if all chunks of the file were read)
func (h *SHA1Hasher) GetFilehash() (string, error) {
	hashInBytes := h.TotalHash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

// HashFile returns the checksum for the specified file. The file is readed completly.
func (h *SHA1Hasher) HashFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)

	return hex.EncodeToString(hashInBytes), nil
}
