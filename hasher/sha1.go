package hasher

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

type SHA1Hasher struct {
	TotalHash hash.Hash
}

func NewSHA1Hasher() *SHA1Hasher {
	h := SHA1Hasher{}
	h.TotalHash = sha1.New()
	return &h
}

func (h *SHA1Hasher) GetName() string {
	return "SHA1"
}

func (h *SHA1Hasher) HashChunk(data []byte) string {
	// write to total hash
	h.TotalHash.Write(data)

	hash := sha1.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func (h *SHA1Hasher) GetFilehash() (string, error) {
	hashInBytes := h.TotalHash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

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
