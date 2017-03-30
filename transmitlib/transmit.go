package transmitlib

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/tsauter/transmit/hasher"
	"github.com/tsauter/transmit/structs"
	"gopkg.in/cheggaaa/pb.v1"
	"net/http"
	"net/url"
	"strconv"
	"time"
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

func CopyHttpToLocal(baseurl string, targetfile string, h *hasher.Hasher, chunksize int) error {
	url, err := url.Parse(baseurl)
	if err != nil {
		return errors.Wrap(err, "invalid url")
	}

	var source SourceFile
	source, err = OpenHttpSource(url)
	if err != nil {
		return errors.Wrap(err, "failed to open local source file")
	}
	defer source.Close()

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

		// for remote file we use the id of the chunk, not the real file position
		data, datalen, err := source.ReadChunkData(int64(chunkStream.ChunkId))
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

func ServeFileOverHttp(listenAddress string, sourcefile string) error {
	var source SourceFile
	source, err := OpenLocalSource(sourcefile)
	if err != nil {
		return errors.Wrap(err, "failed to open local source file")
	}
	defer source.Close()

	fmt.Printf("Loading source cache...\n")
	err = source.LoadCache()
	if err != nil {
		return errors.Wrap(err, "failed to load cache for local source file")
	}

	fileinfo, err := source.GetFileInfo()
	if err != nil {
		return errors.Wrap(err, "failed to get file info for source file")
	}

	r := mux.NewRouter()

	server := &http.Server{
		Addr:         listenAddress,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
	}

	r.HandleFunc("/GetFileInfo", func(w http.ResponseWriter, r *http.Request) {
		jsondata, err := json.Marshal(fileinfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("GetFileInfo: %s\n", err.Error())
			return
		}

		fmt.Printf("Sending file info...\n")
		fmt.Fprintf(w, string(jsondata))
	}).Methods("GET")

	r.HandleFunc("/GetChunk/{chunkno:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		chunkno, err := strconv.ParseUint(mux.Vars(r)["chunkno"], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Printf("GetChunk: %s: %s\n", mux.Vars(r)["chunkno"], err.Error())
			return
		}

		chunk, err := source.GetChunk(chunkno)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("GetChunk: %d: %s\n", chunkno, err.Error())
			return
		}

		jsondata, err := json.Marshal(chunk)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("GetFileInfo: %s\n", err.Error())
			return
		}

		fmt.Printf("Sending chunk...\n")
		fmt.Fprintf(w, string(jsondata))
	}).Methods("GET")

	r.HandleFunc("/GetAllChunks", func(w http.ResponseWriter, r *http.Request) {
		var allChunks []structs.ChunkStream
		_, chunkStreamChan := source.GetAllChunks()
		for chunkStream := range chunkStreamChan {
			allChunks = append(allChunks, chunkStream)
		}

		jsondata, err := json.Marshal(allChunks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("GetAllChunks: %s\n", err.Error())
			return
		}

		fmt.Printf("Sending all chunks...\n")
		fmt.Fprintf(w, string(jsondata))
	}).Methods("GET")

	r.HandleFunc("/ReadChunkData/{chunkno:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		chunkno, err := strconv.ParseUint(mux.Vars(r)["chunkno"], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Printf("ReadChunkData: %s: %s\n", mux.Vars(r)["chunkno"], err.Error())
			return
		}

		chunk, err := source.GetChunk(chunkno)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("ReadChunkData: %d: %s\n", chunkno, err.Error())
			return
		}

		filepos := int64(chunkno * uint64(chunk.Size))

		data, datalen, err := source.ReadChunkData(filepos)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Sending chunk data (%d bytes)...\n", datalen)
		w.Header().Set("Content-Type", "application/octet-stream")
		//w.Header().Set("Content-Length", strconv.Itoa(datalen))
		w.Header().Set("X-ChunkLength", strconv.Itoa(datalen))
		fmt.Fprintf(w, string(data))
	}).Methods("GET")

	fmt.Printf("Waiting for incoming requests...\n")
	err = server.ListenAndServe()
	if err != nil {
		return errors.Wrap(err, "failed to serve file")
	}

	return nil
}
