package hasher

type Hasher interface {
	GetName() string
	HashChunk(data []byte) string
	GetFilehash() (string, error)
	HashFile(filename string) (string, error)
}
