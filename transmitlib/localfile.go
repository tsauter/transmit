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

// LocalFile is the internal representation of the LocalFile
type LocalFile struct {
	// the filename of the file
	filename string
	// the file handle of the file
	f *os.File
	// the hasher that should be used
	h hasher.Hasher
	// chunk size
	chunksize int
	// how and where should we cache the chunks
	cache cache.CacheDB
}

// OpenLocalSource opens the soure file in the local filesystem.
// A LocalFile struct is returned.
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

// OpenOrCreateLocalTarget opens the target file in the filesystem. If the file
// does not exists, it will be created.
// A LocalFile struct is returned.
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

// LoadCache loads the chunk cache database for the local file.
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

// BuildCache regnerates the complete chunk database by rereading the whole file.
// Existing cache data will be removed.
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

// GetFileInfo return the previously stored filedata from the cache database.
func (lf *LocalFile) GetFileInfo() (structs.FileData, error) {
	return lf.cache.GetFileInfo()
}

// SetFilesize resize the file to the specified file size. Unit is bytes.
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

// Close closes the cache database and the open file handle.
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

// CloseAndRemove closes the cache database and the open file handle; the
// cache database will be deleted in the filesystem.
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

// GetAllChunks return all available chunks form database, the chunks are passed
// back through the pipe.
func (lf *LocalFile) GetAllChunks() (int, chan structs.ChunkStream) {
	chunkStreamChan := make(chan structs.ChunkStream, 1)

	numberOfChunks, _ := lf.cache.GetChunksCount()

	go func() {
		lf.cache.GetAllChunks(chunkStreamChan)
		close(chunkStreamChan)
	}()

	return numberOfChunks, chunkStreamChan
}

// ReadChunkData reads the raw data from file (not the chunk) and return the data.
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

// WriteChunkData write the raw data, readed from file, to the file at the specified
// file position.
// The number of bytes to write are specified through datalen. Normally, datalen
// is the chunksize.
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

// GetChunk return the specified chunk details from database.
// This is not the real raw data from file.
func (lf *LocalFile) GetChunk(chunkNo uint64) (structs.Chunk, error) {
	return lf.cache.GetChunk(chunkNo)
}
