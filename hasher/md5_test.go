package hasher

import (
	"path/filepath"
	"testing"
)

func Test_Md5HashChunk(t *testing.T) {
	testcases := []struct {
		Data     string
		Checksum string
	}{
		{"testdata", "ef654c40ab4f1747fc699915d4f70902"},
		{"testdata2", "0e65de7114f9d086a6176fdda0f86e9f"},
	}

	hash := NewMD5Hasher()

	for _, tc := range testcases {
		ressum := hash.HashChunk([]byte(tc.Data))
		if ressum != tc.Checksum {
			t.Errorf("hashing failed: %s: %s", tc.Data, ressum)
		}
	}
}

func Test_Md5HashFile(t *testing.T) {
	testcases := []struct {
		Filename string
		Checksum string
	}{
		{"test1.txt", "4c7b3fc3288e5f9b49138198cc6a8426"},
		{"test2.txt", "cfb789a8e782467d5e9af43f9bb19769"},
	}

	hash := MD5Hasher{}

	for _, tc := range testcases {
		ressum, err := hash.HashFile(filepath.Join("fixtures", tc.Filename))
		if err != nil {
			t.Fatalf("hashing file failed: %s: %s", tc.Filename, err.Error())
		}

		if ressum != tc.Checksum {
			t.Errorf("hashing failed: %s: %s", tc.Filename, ressum)
		}
	}
}
