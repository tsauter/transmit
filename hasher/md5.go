package hasher

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

// Hasher structure to hold internal data.
type MD5Hasher struct {
	TotalHash hash.Hash
}

// NewMD5Hasher returns an initialized MD5 hasher struct.
func NewMD5Hasher() *MD5Hasher {
	h := MD5Hasher{}
	h.TotalHash = md5.New()
	return &h
}

// GetName return the name or hash algorithm of this hasher implementation.
func (h *MD5Hasher) GetName() string {
	return "MD5"
}

// HashChunk takes a byte slice and create a checksum for these bytes.
func (h *MD5Hasher) HashChunk(data []byte) string {
	// write to total hash
	h.TotalHash.Write(data)

	hash := md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// GetFilehash returns the total checksum for all previously processed chunks.
// Usually this is the checksum of the file (if all chunks of the file were read)
func (h *MD5Hasher) GetFilehash() (string, error) {
	hashInBytes := h.TotalHash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

// HashFile returns the checksum for the specified file. The file is readed completly.
func (h *MD5Hasher) HashFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)

	return hex.EncodeToString(hashInBytes), nil
}
