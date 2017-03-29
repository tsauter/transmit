package transmitlib

import (
	"github.com/tsauter/transmit/hasher"
	"github.com/tsauter/transmit/structs"
)

// SourceFile forms an interface that provides all required functions
// the get details and downloads from the source file.
type SourceFile interface {
	// LoadCache loads an existing cache.
	LoadCache() error
	// BuildCache regenerated the complete source cache by reading the whole file.
	BuildCache(h *hasher.Hasher, chunksize int) error
	// GetFileInfo return the stored file information of the source file from cache database.
	GetFileInfo() (structs.FileData, error)
	// GetChunk return the specified chunk details from source database.
	// This is not the real raw data from source file.
	GetChunk(chunkNo uint64) (structs.Chunk, error)
	// GetAllChunks return all available chunks form source database, the chunks are passed
	// back through the pipe.
	GetAllChunks() (int, chan structs.ChunkStream)
	// ReadChunkData reads the raw data from source file (not the chunk) and return the data.
	ReadChunkData(filepos int64) ([]byte, int, error)
	// Close closes the source file and source cache database.
	Close() error
}

// TargetFile forms an interface that provides all required functions
// the get details and write the target file.
type TargetFile interface {
	// BuildCache regenerated the complete target cache by reading the whole file.
	BuildCache(h *hasher.Hasher, chunksize int) error
	// SetFilesize resize the target file to the same size as the source file.
	SetFilesize(newsize int64) error
	// GetChunk return the specified chunk details from target database.
	// This is not the real raw data from target file.
	GetChunk(chunkNo uint64) (structs.Chunk, error)
	// WriteChunkData write the raw data, readed from source file, to the target
	// file at the specified file position.
	// The number of bytes to write are specified through datalen. Normally, datalen
	// is the chunksize.
	WriteChunkData(filepos int64, data []byte, datalen int) error
	// Close closes the target file and target cache database. The target cache database will be
	// removed after closing.
	CloseAndRemove() error
}
