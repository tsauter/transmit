package structs

import (
	"testing"
)

func TestNewChunk(t *testing.T) {
	testcases := []struct {
		Checksum string
		Size     int
	}{
		{"ef654c40ab4f1747fc699915d4f70902", 1000},
		{"0e65de7114f9d086a6176fdda0f86e9f", 1000},
	}

	for _, tc := range testcases {
		chunk := NewChunk(tc.Checksum, tc.Size)

		if chunk.Hash != tc.Checksum {
			t.Errorf("chunk contains different checksum: %s", tc.Checksum)
		}
		if chunk.Size != tc.Size {
			t.Errorf("chunk contains different size: %d", tc.Size)
		}
	}
}
