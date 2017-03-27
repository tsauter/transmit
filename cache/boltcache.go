package cache

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/tsauter/transmit/structs"
	"os"
	"time"
)

const (
	BOLT_BUCKETNAME_INFO   = "info"
	BOLT_BUCKETNAME_CHUNKS = "chunks"
)

type BoltCache struct {
	DbFilename string
	DB         bolt.DB
}

func NewBoltCache() *BoltCache {
	bc := &BoltCache{}
	return bc
}

// InitDatabase creates a new BoltDB database and initialize some default buckets
// The filename of the database is specified in the cachefilename parameter
func (bc *BoltCache) InitDatabase(cachefilename string) error {
	cachefilename = cachefilename + ".db" // the .db is required for Bolt databases
	bc.DbFilename = cachefilename

	db, err := bolt.Open(cachefilename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return errors.Wrapf(err, "failed to create cache database (%s)", cachefilename)
	}
	bc.DB = *db

	// create the bolt bucket, that holds the file details
	err = bc.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BOLT_BUCKETNAME_INFO))
		return err
	})
	if err != nil {
		return errors.Wrap(err, "updating chunk database failed")
	}

	// create the bolt bucket, that holds the chunks
	err = bc.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BOLT_BUCKETNAME_CHUNKS))
		return err
	})
	if err != nil {
		return errors.Wrap(err, "updating chunk database failed")
	}

	return nil
}

// CloseDatabase sync and close the bolt database
func (bc *BoltCache) CloseDatabase() error {
	err := bc.DB.Close()
	return err
}

// Cleanup delete the bolt db file in the filesystem
func (bc *BoltCache) Cleanup() error {
	// make sure the db is already closed
	err := bc.CloseDatabase()
	if err != nil {
		return err
	}

	// delete the file
	err = os.Remove(bc.DbFilename)
	if err != nil {
		return err
	}

	return nil
}

// ClearAllChunks remove all previous stored chunks from bolt database
// (delete * from chunks)
func (bc *BoltCache) ClearAllChunks() error {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(BOLT_BUCKETNAME_CHUNKS))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucket([]byte(BOLT_BUCKETNAME_CHUNKS))
		return err
	})
	if err != nil {
		return errors.Wrap(err, "clear chunks failed")
	}
	return nil
}

// GetFileInfo reads the stored file information from the bolt database.
// The file information will be returned in a FileData struct and an optional
// error might be returned.
// The file information is stored as marshaled json data in the bucket BOLT_BUCKETNAME_INFO
func (bc *BoltCache) GetFileInfo() (structs.FileData, error) {
	fd := structs.FileData{}
	var err error

	var jsonbytes []byte

	err = bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BOLT_BUCKETNAME_INFO))
		jsonbytes = b.Get([]byte(BOLT_BUCKETNAME_INFO))
		return nil
	})
	if err != nil {
		return fd, err
	}

	err = json.Unmarshal(jsonbytes, &fd)
	return fd, err
}

// StoreFileInfo takes a FileData struct and store those data
// in the bolt database.
// If an error occures, this error will be returned.
// The file information is stored as marshaled json data in the bucket BOLT_BUCKETNAME_INFO
func (bc *BoltCache) StoreFileInfo(fd structs.FileData) error {
	marshaled_data, err := json.Marshal(fd)
	if err != nil {
		return err
	}

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BOLT_BUCKETNAME_INFO))
		return b.Put([]byte(BOLT_BUCKETNAME_INFO), marshaled_data)
	})
	return err
}

// GetChunk returns the chunk stored under the paramter chunkid.
// An error is return when the chunk was not found.
// The complete Chunk strict will be returned.
func (bc *BoltCache) GetChunk(chunkId uint64) (structs.Chunk, error) {
	var jsonbytes []byte
	var chunk structs.Chunk

	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BOLT_BUCKETNAME_CHUNKS))
		jsonbytes = b.Get(itob(chunkId))
		return nil
	})

	// empty value? chunk not found
	if jsonbytes == nil {
		return chunk, fmt.Errorf("chunk %d not found", chunkId)
	}

	// unmarshal the data to a chunk struct
	err = json.Unmarshal(jsonbytes, &chunk)
	return chunk, err
}

// StoreChunk store the passed chunk in the database. The chunk,
// is a marshalled json string.
func (bc *BoltCache) StoreChunk(chunkId uint64, chunk structs.Chunk) error {
	marshaled_data, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BOLT_BUCKETNAME_CHUNKS))
		return b.Put(itob(chunkId), marshaled_data)
	})

	return err
}

// GetChunksCount return the number of stored chunks.
// In case of an error, this error is returned.
func (bc *BoltCache) GetChunksCount() (int, error) {
	var stats bolt.BucketStats

	err := bc.DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(BOLT_BUCKETNAME_CHUNKS))
		stats = b.Stats()
		return nil
	})
	if err != nil {
		return 0, err
	}

	return stats.KeyN, err
}

// GetChunksCount return the number of stored chunks.
// In case of an error, this error is returned.
// TODO: do we need to sort them?
func (bc *BoltCache) GetAllChunks(chunkStreamChan chan structs.ChunkStream) error {
	err := bc.DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(BOLT_BUCKETNAME_CHUNKS))

		c := b.Cursor()

		i := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("    %d: %s\n", i, v)
			var chunk structs.Chunk
			err := json.Unmarshal(v, &chunk)
			if err != nil {
				return err
			}

			pos, err := btoi(k)
			if err != nil {
				return fmt.Errorf("Failed to convert byte to uint64")
			}

			chunkstrm := structs.ChunkStream{ChunkId: pos, Chunk: chunk}

			chunkStreamChan <- chunkstrm
			i++
		}

		return nil
	})
	return err
}

// itab converts a uint64 value into an 8 byte long byte slice.
// This byte slice can be used as a key value for the bolt database.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(b []byte) (uint64, error) {
	var i uint64
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}
