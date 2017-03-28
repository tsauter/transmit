package structs

// FileData contains file details for a unique file.
// Usually this information is stored in the cache database to indentify
// the source file of the database.
type FileData struct {
	// The filename of the file, without directory
	Filename string `json:"filename"`
	// File size in bytes
	Filesize int64 `json:"filesize"`
	// The checksum, format depends on the used hasher
	Checksum string `json:"checksum"`
	// The used hash algorithm as string, depends on the used hasher
	ChunkHashAlgorithm string `json:"hashalgo"`
	// The default size of all chunks, this can be overwritten by each individual chunk
	Chunksize int `jons:"chunksize"`
}
