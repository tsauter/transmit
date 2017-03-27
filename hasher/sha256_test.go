package hasher

import (
	"path/filepath"
	"testing"
)

func Test_Sha256HashChunk(t *testing.T) {
	testcases := []struct {
		Data     string
		Checksum string
	}{
		{"testdata", "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"},
		{"testdata2", "80905dfa0896b4f02b65e7fc3f7244e2f20bf7346191db14fa66f0645790244d"},
	}

	hash := NewSHA256Hasher()

	for _, tc := range testcases {
		ressum := hash.HashChunk([]byte(tc.Data))
		if ressum != tc.Checksum {
			t.Errorf("hashing failed: %s: %s", tc.Data, ressum)
		}
	}
}

func Test_Sha256HashFile(t *testing.T) {
	testcases := []struct {
		Filename string
		Checksum string
	}{
		{"test1.txt", "b183917121a41b7d24bc984edb330eb4ebcf9af3e9263cc619fbebcc20c185d7"},
		{"test2.txt", "1966a2dcacfce86969007d91b766e13615f9a6d2bb674dd3504e2f5b97c54720"},
	}

	hash := SHA256Hasher{}

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
