package cache

import (
	"github.com/tsauter/transmit/structs"
)

// CacheDB is the generic interface for chunk cache backends.
// Backends could be bolt, mysql, json...
type CacheDB interface {
	// Open a connecton to the database. This will not fill chunk details.
	InitDatabase(sourcefile string) error
	// Close the connection to the database
	CloseDatabase() error
	// Cleanup database (delete table or file; depending on the implementation)
	Cleanup() error

	// Remove all existing chunks from database
	ClearAllChunks() error

	// Return stored file details from cache database
	GetFileInfo() (structs.FileData, error)
	// Store file details in cache database
	StoreFileInfo(fd structs.FileData) error

	// Get individual chunks (by chunk id) from database
	GetChunk(chunkId uint64) (structs.Chunk, error)
	// Store chunk under the specified chunk id
	StoreChunk(chunkId uint64, chunk structs.Chunk) error

	// Get the total number of stored chunks
	GetChunksCount() (int, error)
	// Return a channel to iterate over all stored chunks
	GetAllChunks(chunkChan chan structs.ChunkStream) error
}
