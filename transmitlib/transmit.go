package transmitlib

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tsauter/transmit/hasher"
	"gopkg.in/cheggaaa/pb.v1"
)

// CopyLocalToLocal copy the sourcefile to the targetfile. Both files must be available on
// the local filesystem (drives, mount points, mounted nfs or smb shares).
// The hasher and chunksize parameter must must the options used in the source file cache.
func CopyLocalToLocal(sourcefile string, targetfile string, h *hasher.Hasher, chunksize int) error {
	var source SourceFile
	source, err := OpenLocalSource(sourcefile)
	if err != nil {
		return errors.Wrap(err, "failed to open local source file")
	}
	defer source.Close()

	fmt.Printf("Loading source cache...\n")
	err = source.LoadCache()
	if err != nil {
		return errors.Wrap(err, "failed to local cache for local source file")
	}

	var targethasher hasher.Hasher = *h
	var target TargetFile
	target, err = OpenOrCreateLocalTarget(targetfile)
	if err != nil {
		return errors.Wrap(err, "failed to open target file")
	}
	defer target.CloseAndRemove()

	sourceinfo, err := source.GetFileInfo()
	if err != nil {
		return errors.Wrap(err, "failed to get file info for source file")
	}

	err = target.SetFilesize(sourceinfo.Filesize)
	if err != nil {
		return errors.Wrap(err, "unable to resize target file to new filesize")
	}

	fmt.Printf("Building local file cache...\n")
	err = target.BuildCache(&targethasher, chunksize)
	if err != nil {
		return errors.Wrap(err, "failed to build cache for local target file")
	}

	// walk over the list of stored source chunks,
	// compaire the chunk checksum with the target checksum
	// read/write chunk data if both hashes missmatch
	fmt.Printf("Copy individual file chunks...\n")
	maxchunkno, chunkStreamChan := source.GetAllChunks()
	percentBar := pb.StartNew(int(maxchunkno) + 1)
	for chunkStream := range chunkStreamChan {
		percentBar.Increment()

		dstchunk, err := target.GetChunk(chunkStream.ChunkId)
		if err != nil {
			return errors.Wrapf(err, "failed to get chunk from target: %d: %s", chunkStream.ChunkId, err.Error())
		}

		// comparing both chunks, do nothing if both are equal
		if chunkStream.Chunk.Hash == dstchunk.Hash {
			//fmt.Printf("Chunk %d is equal: %s == %s\n", chunkStream.ChunkId, chunkStream.Chunk.Hash, dstchunk.Hash)
			continue
		}
		//fmt.Printf("Chunk %d different: %s != %s\n", chunkStream.ChunkId, chunkStream.Chunk.Hash, dstchunk.Hash)

		// fetch the required data from files

		filepos := int64(chunkStream.ChunkId * uint64(chunksize))

		data, datalen, err := source.ReadChunkData(filepos)
		if err != nil {
			return errors.Wrapf(err, "failed to read from source: %s", err.Error())
		}

		err = target.WriteChunkData(filepos, data, datalen)
		if err != nil {
			return errors.Wrapf(err, "failed to write to target: %s", err.Error())
		}
	}
	percentBar.FinishPrint("Finish.")

	fmt.Printf("Validating checksum...\n")
	tchecksum, err := targethasher.HashFile(targetfile)
	if err != nil {
		return errors.Wrapf(err, "failed to calculate checksum: %s: %s", targetfile, err.Error())
	}
	if sourceinfo.Checksum != tchecksum {
		return errors.Wrapf(err, "checksum is different, test returns different data")
	}

	return nil
}
