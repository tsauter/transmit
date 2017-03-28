package hasher

// Interface for building checksums for bytes slices.
type Hasher interface {
	GetName() string
	HashChunk(data []byte) string
	GetFilehash() (string, error)
	HashFile(filename string) (string, error)
}
