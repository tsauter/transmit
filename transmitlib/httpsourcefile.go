package transmitlib

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tsauter/transmit/hasher"
	"github.com/tsauter/transmit/structs"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// HttpFile is the internal representation of the HttpFile
type HttpFile struct {
	// the filename of the file
	baseUrl    *url.URL
	httpclient *http.Client
}

// OpenLocalHttpSource opens the soure file in the local filesystem.
// A HttpFile struct is returned.
func OpenHttpSource(url *url.URL) (*HttpFile, error) {
	hf := HttpFile{baseUrl: url}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false, //TODO:???
	}

	hf.httpclient = &http.Client{Transport: tr}

	return &hf, nil
}

// LoadCache loads the chunk cache database for the local file.
func (hf *HttpFile) LoadCache() error {
	// loading a remote cache is not necessary
	return nil
}

// BuildCache regnerates the complete chunk database by rereading the whole file.
// Existing cache data will be removed.
func (hf *HttpFile) BuildCache(h *hasher.Hasher, chunksize int) error {
	return fmt.Errorf("remote building of cache is not possible")
}

// GetFileInfo return the previously stored filedata from the cache database.
func (hf *HttpFile) GetFileInfo() (structs.FileData, error) {
	content, err := hf.FetchRemoteBytes("GetFileInfo")

	var data structs.FileData
	err = json.Unmarshal(content, &data)
	if err != nil {
		return structs.FileData{}, errors.Wrap(err, "failed to read file info from remote server")
	}

	return data, nil
}

// Close closes the cache database and the open file handle.
func (hf *HttpFile) Close() error {
	// nothing to close for remote connections
	return nil
}

// GetAllChunks return all available chunks form database, the chunks are passed
// back through the pipe.
func (hf *HttpFile) GetAllChunks() (int, chan structs.ChunkStream) {
	chunkStreamChan := make(chan structs.ChunkStream, 1)

	content, err := hf.FetchRemoteBytes("GetAllChunks")

	var data []structs.ChunkStream
	err = json.Unmarshal(content, &data)
	if err != nil {
		close(chunkStreamChan)
		panic(err)
		return 0, chunkStreamChan // TODO: how can we transmit errors here?
	}
	numberOfChunks := len(data)

	go func() {
		for _, stream := range data {
			chunkStreamChan <- stream
		}

		close(chunkStreamChan)
	}()

	return numberOfChunks, chunkStreamChan
}

// ReadChunkData reads the raw data from file (not the chunk) and return the data.
func (hf *HttpFile) ReadChunkData(filepos int64) ([]byte, int, error) {
	buf, err := hf.FetchRemoteBytes(fmt.Sprintf("ReadChunkData/%d", filepos))
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to read remote chunk data")
	}

	return buf, len(buf), nil
}

// GetChunk return the specified chunk details from database.
// This is not the real raw data from file.
func (hf *HttpFile) GetChunk(chunkNo uint64) (structs.Chunk, error) {
	content, err := hf.FetchRemoteBytes(fmt.Sprintf("GetChunk/%d", chunkNo))

	var data structs.Chunk
	err = json.Unmarshal(content, &data)
	if err != nil {
		return structs.Chunk{}, errors.Wrap(err, "failed to read chunk from remote server")
	}

	return data, nil
}

func (hf *HttpFile) BuildRequestUrl(method string) string {
	return hf.baseUrl.String() + "/" + method
}

func (hf *HttpFile) FetchRemoteBytes(method string) ([]byte, error) {
	resp, err := hf.httpclient.Get(hf.BuildRequestUrl(method))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get data from remote server")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to query remote size: %d: %s", resp.StatusCode, resp.Request.URL.String())
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read data from remote server")
	}

	if resp.Header.Get("X-Chunklength") != "" {
		length, err := strconv.Atoi(resp.Header.Get("X-Chunklength"))
		if err == nil {
			content = content[:length]
		}
	}

	return content, nil
}
