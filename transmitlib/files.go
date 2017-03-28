package transmitlib

import (
	"github.com/tsauter/transmit/hasher"
	"github.com/tsauter/transmit/structs"
)

type SourceFile interface {
	LoadCache() error
	BuildCache(h *hasher.Hasher, chunksize int) error
	GetFileInfo() (structs.FileData, error)
	GetChunk(chunkNo uint64) (structs.Chunk, error)
	GetAllChunks() (int, chan structs.ChunkStream)
	ReadChunkData(filepos int64) ([]byte, int, error)
	Close() error
}

type TargetFile interface {
	BuildCache(h *hasher.Hasher, chunksize int) error
	SetFilesize(newsize int64) error
	GetChunk(chunkNo uint64) (structs.Chunk, error)
	WriteChunkData(filepos int64, data []byte, datalen int) error
	CloseAndRemove() error
}
