package structs

type FileData struct {
	Filename           string `json:"filename"`
	Filesize           int64  `json:"filesize"`
	Checksum           string `json:"checksum"`
	ChunkHashAlgorithm string `json:"hashalgo"`
	Chunksize          int    `jons:"chunksize"`
}
