package structs

// This struct represents the individual chunk.
type Chunk struct {
	// The checksum of this chunk, depending on the used hasher.
	Hash string `json:"hash"`
	// The size of the chunk, can be overwritten with chunksize.
	// TODO: can we avoid the size parameter here=
	Size int `json:"size,omitempty"`
}

// ChunkStream contains the chunk id/position and the Chunk details itself.
// It is a combination between id and data.
type ChunkStream struct {
	ChunkId uint64
	Chunk   Chunk
}

// NewChunk create a new Chunk struct filled with the arguments.
func NewChunk(checksum string, size int) Chunk {
	return Chunk{Hash: checksum, Size: size}
}
