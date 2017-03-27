package hasher

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

type MD5Hasher struct {
	TotalHash hash.Hash
}

func NewMD5Hasher() *MD5Hasher {
	h := MD5Hasher{}
	h.TotalHash = md5.New()
	return &h
}

func (h *MD5Hasher) GetName() string {
	return "MD5"
}

func (h *MD5Hasher) HashChunk(data []byte) string {
	// write to total hash
	h.TotalHash.Write(data)

	hash := md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func (h *MD5Hasher) GetFilehash() (string, error) {
	hashInBytes := h.TotalHash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

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
