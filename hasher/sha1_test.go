package hasher

import (
	"path/filepath"
	"testing"
)

func Test_Sha1HashChunk(t *testing.T) {
	testcases := []struct {
		Data     string
		Checksum string
	}{
		{"testdata", "44115646e09ab3481adc2b1dc17be10dd9cdaa09"},
		{"testdata2", "aabd7d1665facce1b48b56a111ab82818e0068df"},
	}

	hash := NewSHA1Hasher()

	for _, tc := range testcases {
		ressum := hash.HashChunk([]byte(tc.Data))
		if ressum != tc.Checksum {
			t.Errorf("hashing failed: %s: %s", tc.Data, ressum)
		}
	}
}

func Test_Sha1HashFile(t *testing.T) {
	testcases := []struct {
		Filename string
		Checksum string
	}{
		{"test1.txt", "bbe707bba45498134129639a06f2a9d6b2f22c5e"},
		{"test2.txt", "f01e4c9d79f29feff3346a1bfc72f858c4cf5ba4"},
	}

	hash := SHA1Hasher{}

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
