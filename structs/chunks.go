package structs

type Chunk struct {
	Hash string `json:"hash"`
	// TOOD:can we avoid the size parameter here=
	Size int `json:"size,omitempty"`
}

// ChunkStream contains the chunk id/position and the Chunk details itself.
// It is a combination between id and data.
type ChunkStream struct {
	ChunkId uint64
	Chunk   Chunk
}

func NewChunk(checksum string, size int) Chunk {
	return Chunk{Hash: checksum, Size: size}
}
