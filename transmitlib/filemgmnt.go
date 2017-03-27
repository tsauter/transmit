package transmitlib

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tsauter/transmit/cache"
	"github.com/tsauter/transmit/hasher"
	"github.com/tsauter/transmit/structs"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"os"
	"path/filepath"
	"strings"
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

type LocalFile struct {
	filename  string
	f         *os.File
	h         hasher.Hasher
	chunksize int
	cache     cache.CacheDB
}

func OpenLocalSource(filename string) (*LocalFile, error) {
	lf := LocalFile{filename: filename}

	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open or create file")
	}
	lf.f = f

	// we use BoltDB for caching
	lf.cache = cache.NewBoltCache()

	return &lf, nil
}

func OpenOrCreateLocalTarget(filename string) (*LocalFile, error) {
	lf := LocalFile{filename: filename}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 644)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open or create file")
	}
	lf.f = f

	// we use BoltDB for caching
	lf.cache = cache.NewBoltCache()

	return &lf, nil
}

func (lf *LocalFile) LoadCache() error {
	// read the file
	err := lf.cache.InitDatabase(lf.filename + ".tcache")
	if err != nil {
		return errors.Wrap(err, "failed to open or create file")
	}

	info, err := lf.cache.GetFileInfo()
	if err != nil {
		return errors.Wrap(err, "failed to load file info from cache")
	}

	lf.chunksize = info.Chunksize

	var h hasher.Hasher
	switch strings.ToLower(info.ChunkHashAlgorithm) {
	case "sha256":
		h = hasher.NewSHA256Hasher()
	case "sha1":
		h = hasher.NewSHA1Hasher()
	case "md5":
		h = hasher.NewMD5Hasher()
	default:
		return fmt.Errorf("unsupported hash algorithm: %s\n", info.ChunkHashAlgorithm)
	}
	lf.h = h

	return nil
}

func (lf *LocalFile) BuildCache(h *hasher.Hasher, chunksize int) error {
	lf.chunksize = chunksize
	lf.h = *h

	// make sure chunksize is greater than 0
	if chunksize < 1 {
		return fmt.Errorf("chunksize %d to small", chunksize)
	}

	// read the file
	err := lf.cache.InitDatabase(lf.filename + ".tcache")
	if err != nil {
		return errors.Wrap(err, "failed to open or create file")
	}

	// remove all pre existing chunks in database
	err = lf.cache.ClearAllChunks()
	if err != nil {
		return errors.Wrap(err, "failed to clear existing chunks")
	}

	// get file info from operating system
	fstat, err := lf.f.Stat()
	if err != nil {
		return errors.Wrapf(err, "failed to get file info")
	}

	fd := structs.FileData{}
	fd.Filename = filepath.Base(lf.filename)
	fd.Filesize = fstat.Size()
	fd.ChunkHashAlgorithm = lf.h.GetName()
	fd.Chunksize = lf.chunksize

	maxchunkno := fd.Filesize / int64(lf.chunksize)
	percentBar := pb.StartNew(int(maxchunkno) + 1)

	var chunkno uint64 = 0
	buf := make([]byte, lf.chunksize)
	for {
		n, err := lf.f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrapf(err, "failed to read chunk %d from file %s", chunkno, lf.filename)
		}

		chunk := structs.NewChunk(lf.h.HashChunk(buf[:n]), len(buf[:n]))
		lf.cache.StoreChunk(chunkno, chunk)

		percentBar.Increment()

		chunkno++
	}
	percentBar.FinishPrint("Finish.")

	// return the checksum of the complete file
	checksum, err := lf.h.GetFilehash()
	if err != nil {
		return errors.Wrapf(err, "failed to create checksum for file %s", lf.filename)
	}
	fd.Checksum = checksum

	err = lf.cache.StoreFileInfo(fd)
	if err != nil {
		return errors.Wrapf(err, "failed to store checksum for file %s", lf.filename)
	}

	return nil
}

func (lf *LocalFile) GetFileInfo() (structs.FileData, error) {
	return lf.cache.GetFileInfo()
}

func (lf *LocalFile) SetFilesize(newsize int64) error {
	stats, err := lf.f.Stat()
	if err != nil {
		return errors.Wrap(err, "failed to get filesize")
	}
	myfilesize := stats.Size()

	// shrink or extend the file to the specified filesize
	if myfilesize == newsize {
		return nil
	}

	// file size is different, truncate the file to the expected size
	err = lf.f.Truncate(newsize)
	if err != nil {
		return errors.Wrap(err, "failed to resize file")
	}

	return nil
}

func (lf *LocalFile) Close() error {
	// close the cache database
	if lf.cache != nil {
		err := lf.cache.CloseDatabase()
		if err != nil {
			return errors.Wrap(err, "failed to close cache")
		}
	}

	// close the open file handle
	err := lf.f.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close file")
	}

	return nil
}

func (lf *LocalFile) CloseAndRemove() error {
	// close the cache database
	if lf.cache != nil {
		err := lf.cache.CloseDatabase()
		if err != nil {
			return errors.Wrap(err, "failed to close cache")
		}
	}

	// close the open file handle
	err := lf.f.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close file")
	}

	// delete the cache database
	lf.cache.Cleanup()

	return nil
}

func (lf *LocalFile) GetAllChunks() (int, chan structs.ChunkStream) {
	chunkStreamChan := make(chan structs.ChunkStream, 1)

	numberOfChunks, _ := lf.cache.GetChunksCount()

	go func() {
		lf.cache.GetAllChunks(chunkStreamChan)
		close(chunkStreamChan)
	}()

	return numberOfChunks, chunkStreamChan
}

func (lf *LocalFile) ReadChunkData(filepos int64) ([]byte, int, error) {
	_, err := lf.f.Seek(filepos, 0)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to seek file")
	}

	buf := make([]byte, lf.chunksize)
	buflen, err := lf.f.Read(buf)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to read file")
	}
	return buf, buflen, nil
}

func (lf *LocalFile) WriteChunkData(filepos int64, data []byte, datalen int) error {
	_, err := lf.f.Seek(filepos, 0)
	if err != nil {
		return errors.Wrap(err, "failed to seek file")
	}

	_, err = lf.f.Write(data[:datalen])
	if err != nil {
		return errors.Wrap(err, "failed to write chunk to file")
	}

	return nil
}

func (lf *LocalFile) GetChunk(chunkNo uint64) (structs.Chunk, error) {
	return lf.cache.GetChunk(chunkNo)
}
