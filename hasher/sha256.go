package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

type SHA256Hasher struct {
	TotalHash hash.Hash
}

func NewSHA256Hasher() *SHA256Hasher {
	h := SHA256Hasher{}
	h.TotalHash = sha256.New()
	return &h
}

func (h *SHA256Hasher) GetName() string {
	return "SHA256"
}

func (h *SHA256Hasher) HashChunk(data []byte) string {
	// write to total hash
	h.TotalHash.Write(data)

	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func (h *SHA256Hasher) GetFilehash() (string, error) {
	hashInBytes := h.TotalHash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

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
