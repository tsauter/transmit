package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

// Hasher structure to hold internal data.
type SHA256Hasher struct {
	TotalHash hash.Hash
}

// NewSHA256Hasher returns an initialized SHA256 hasher struct.
func NewSHA256Hasher() *SHA256Hasher {
	h := SHA256Hasher{}
	h.TotalHash = sha256.New()
	return &h
}

// GetName return the name or hash algorithm of this hasher implementation.
func (h *SHA256Hasher) GetName() string {
	return "SHA256"
}

// HashChunk takes a byte slice and create a checksum for these bytes.
func (h *SHA256Hasher) HashChunk(data []byte) string {
	// write to total hash
	h.TotalHash.Write(data)

	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// GetFilehash returns the total checksum for all previously processed chunks.
// Usually this is the checksum of the file (if all chunks of the file were read)
func (h *SHA256Hasher) GetFilehash() (string, error) {
	hashInBytes := h.TotalHash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

// HashFile returns the checksum for the specified file. The file is readed completly.
func (h *SHA256Hasher) HashFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)

	return hex.EncodeToString(hashInBytes), nil
}
