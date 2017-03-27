package cache

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/tsauter/transmit/structs"
	"os"
	"reflect"
	"testing"
)

func TestInitClose(t *testing.T) {
	// create and initialize the database
	boltcache := NewBoltCache()
	err := boltcache.InitDatabase("gotest.cache")
	if err != nil {
		t.Errorf("Fail to create database: %s", err.Error())
	}

	// make sure all required top level buckets exist after the
	// initialization
	// create the bolt bucket, that holds the file details
	requiredBuckets := []string{BOLT_BUCKETNAME_INFO, BOLT_BUCKETNAME_CHUNKS}
	for _, bucketname := range requiredBuckets {
		err = boltcache.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucketname))
			if b == nil {
				return fmt.Errorf("Bucket %s not exists.", bucketname)
			}
			return nil
		})
		if err != nil {
			t.Errorf("Missing bucket: %s", err.Error())
		}
	}

	// close and delete the database, in case of deleting is not
	// working, the database was not properly closed
	err = boltcache.CloseDatabase()
	if err != nil {
		t.Errorf("Fail to close database: %s", err.Error())
	}
	err = os.Remove(boltcache.DbFilename)
	if err != nil {
		t.Errorf("Fail to delete database file %s: DB not closed: %s", boltcache.DbFilename, err.Error())
	}
}

func TestGetStoreFileInfo(t *testing.T) {
	testcases := []struct {
		Name string
		Data structs.FileData
	}{
		{
			Name: "case1",
			Data: structs.FileData{
				Filename:           "mytestfile.txt",
				Filesize:           1024,
				Checksum:           "5ce1a1b956e5336e8a509f4b794f446bbbfec818",
				ChunkHashAlgorithm: "SHA1",
				Chunksize:          1024,
			},
		},
		{
			Name: "case2",
			Data: structs.FileData{
				Filename:           "large.iso",
				Filesize:           202020202,
				Checksum:           "9940b28d7ec4fcd6cbaa3333a4c3db4c31692d03",
				ChunkHashAlgorithm: "SHA1",
				Chunksize:          348728,
			},
		},
	}

	for _, tc := range testcases {
		// create and initialize the database
		boltcache := NewBoltCache()
		err := boltcache.InitDatabase("gotest.cache")
		if err != nil {
			t.Errorf("Fail to create database: %s", err.Error())
		}

		// store the FileData variable in the info bucket
		boltcache.StoreFileInfo(tc.Data)

		// read the value
		fd2, err := boltcache.GetFileInfo()
		if err != nil {
			t.Errorf("Fail to get file info: %s", err.Error())
		}

		// compare both variables
		if !reflect.DeepEqual(tc.Data, fd2) {
			t.Errorf("FileInfo data is not equal. Missmatch between storing and getting.")
		}

		// close the database and reopen again, and read the value again
		// this makes sure, that the value is really written to disk
		err = boltcache.CloseDatabase()
		if err != nil {
			t.Errorf("Fail to close database: %s", err.Error())
		}
		err = boltcache.InitDatabase("gotest.cache")
		if err != nil {
			t.Errorf("Fail to create database: %s", err.Error())
		}
		fd2, err = boltcache.GetFileInfo()
		if err != nil {
			t.Errorf("Fail to get file info: %s", err.Error())
		}
		if !reflect.DeepEqual(tc.Data, fd2) {
			t.Errorf("FileInfo data is not equal. Missmatch between storing and getting.")
		}

		// close and delete the database, in case of deleting is not
		// working, the database was not properly closed
		err = boltcache.CloseDatabase()
		if err != nil {
			t.Errorf("Fail to close database: %s", err.Error())
		}
		err = os.Remove(boltcache.DbFilename)
		if err != nil {
			t.Errorf("Fail to delete database file %s: DB not closed: %s", boltcache.DbFilename, err.Error())
		}

	}
}

func TestGetStoreChunks(t *testing.T) {
	testcases := []struct {
		Name   string
		Chunks []structs.Chunk
	}{
		{
			Name: "case1",
			Chunks: []structs.Chunk{
				{Hash: "case1hash1"},
				{Hash: "case1hash2"},
				{Hash: "case1hash3"},
				{Hash: "case1hash4"},
				{Hash: "case1hash5"},
			},
		},
		{
			Name: "case2",
			Chunks: []structs.Chunk{
				{Hash: "case2hash1"},
				{Hash: "case2hash2"},
				{Hash: "case2hash3"},
				{Hash: "case2hash4"},
				{Hash: "case2hash5"},
			},
		},
	}

	for _, tc := range testcases {
		// create and initialize the database
		boltcache := NewBoltCache()
		err := boltcache.InitDatabase("gotest.cache")
		if err != nil {
			t.Errorf("Fail to create database: %s", err.Error())
		}

		// store all chunks
		for pos, chunk := range tc.Chunks {
			err = boltcache.StoreChunk(uint64(pos), chunk)
			if err != nil {
				t.Errorf("Fail to store chunk: %s", err.Error())
			}
		}

		// read all chunks
		for pos, chunk := range tc.Chunks {
			chunk2, err := boltcache.GetChunk(uint64(pos))
			if err != nil {
				t.Errorf("Fail to read chunk: %s", err.Error())
			}

			// compare both variables
			if !reflect.DeepEqual(chunk, chunk2) {
				t.Errorf("Retrieved chunk data is not equal. Missmatch between storing and getting.")
			}

		}

		// close and delete the database, in case of deleting is not
		// working, the database was not properly closed
		err = boltcache.CloseDatabase()
		if err != nil {
			t.Errorf("Fail to close database: %s", err.Error())
		}
		err = os.Remove(boltcache.DbFilename)
		if err != nil {
			t.Errorf("Fail to delete database file %s: DB not closed: %s", boltcache.DbFilename, err.Error())
		}

	}
}

func TestGetChunksCount(t *testing.T) {
	testcases := []struct {
		Name   string
		Chunks []structs.Chunk
	}{
		{
			Name: "case1",
			Chunks: []structs.Chunk{
				{Hash: "case1hash1"},
			},
		},
		{
			Name: "case2",
			Chunks: []structs.Chunk{
				{Hash: "case2hash1"},
				{Hash: "case2hash2"},
			},
		},
		{
			Name: "case3",
			Chunks: []structs.Chunk{
				{Hash: "case3hash1"},
				{Hash: "case3hash2"},
				{Hash: "case3hash3"},
			},
		},
	}

	for _, tc := range testcases {
		// create and initialize the database
		boltcache := NewBoltCache()
		err := boltcache.InitDatabase("gotest.cache")
		if err != nil {
			t.Errorf("Fail to create database: %s", err.Error())
		}

		// store all chunks
		for pos, chunk := range tc.Chunks {
			err = boltcache.StoreChunk(uint64(pos), chunk)
			if err != nil {
				t.Errorf("Fail to store chunk: %s", err.Error())
			}
		}

		// return the number of chunks and compare them
		count, err := boltcache.GetChunksCount()
		if err != nil {
			t.Errorf("Fail to get count of stored chunks: %s", err.Error())
		}
		if count != len(tc.Chunks) {
			t.Errorf("Invalid count returned: %d", count)
		}

		// close and delete the database, in case of deleting is not
		// working, the database was not properly closed
		err = boltcache.CloseDatabase()
		if err != nil {
			t.Errorf("Fail to close database: %s", err.Error())
		}
		err = os.Remove(boltcache.DbFilename)
		if err != nil {
			t.Errorf("Fail to delete database file %s: DB not closed: %s", boltcache.DbFilename, err.Error())
		}

	}
}

func TestGetAllChunks(t *testing.T) {
	testcases := []struct {
		Name   string
		Chunks []structs.Chunk
	}{
		{
			Name: "case1",
			Chunks: []structs.Chunk{
				{Hash: "case1hash1"},
			},
		},
		{
			Name: "case2",
			Chunks: []structs.Chunk{
				{Hash: "case2hash1"},
				{Hash: "case2hash2"},
			},
		},
		{
			Name: "case3",
			Chunks: []structs.Chunk{
				{Hash: "case3hash1"},
				{Hash: "case3hash2"},
				{Hash: "case3hash3"},
			},
		},
	}

	for _, tc := range testcases {
		// create and initialize the database
		boltcache := NewBoltCache()
		err := boltcache.InitDatabase("gotest.cache")
		if err != nil {
			t.Errorf("Fail to create database: %s", err.Error())
		}

		// store all chunks
		for pos, chunk := range tc.Chunks {
			err = boltcache.StoreChunk(uint64(pos), chunk)
			if err != nil {
				t.Errorf("Fail to store chunk: %s", err.Error())
			}
		}

		// start a new background go routine that iterates of all available chunks
		chunkStreamChan := make(chan structs.ChunkStream)
		go func() {
			err = boltcache.GetAllChunks(chunkStreamChan)
			if err != nil {
				t.Errorf("Fail to walk over all chunks: %s", err.Error())
			}
			close(chunkStreamChan)
		}()

		var tmpchunklist []structs.Chunk
		for chunkstrm := range chunkStreamChan {
			tmpchunklist = append(tmpchunklist, chunkstrm.Chunk)
		}

		if !reflect.DeepEqual(tc.Chunks, tmpchunklist) {
			t.Errorf("Returned list of chunks is different.")
		}

		// close and delete the database, in case of deleting is not
		// working, the database was not properly closed
		err = boltcache.CloseDatabase()
		if err != nil {
			t.Errorf("Fail to close database: %s", err.Error())
		}
		err = os.Remove(boltcache.DbFilename)
		if err != nil {
			t.Errorf("Fail to delete database file %s: DB not closed: %s", boltcache.DbFilename, err.Error())
		}

	}
}
