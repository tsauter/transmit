package transmitlib

import (
	"flag"
	"fmt"
	"github.com/tsauter/transmit/hasher"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	testcases = []struct {
		filename  string
		hasher    hasher.Hasher
		chunksize int
	}{
		{
			filename:  "test0.txt",
			hasher:    hasher.NewSHA256Hasher(),
			chunksize: 2,
		},
		{
			filename:  "test1.txt",
			hasher:    hasher.NewSHA256Hasher(),
			chunksize: 2,
		},
		{
			filename:  "test2.txt",
			hasher:    hasher.NewMD5Hasher(),
			chunksize: 2,
		},
		{
			filename:  "test3.txt",
			hasher:    hasher.NewMD5Hasher(),
			chunksize: 4,
		},
		{
			//https://cdn.kernel.org/pub/linux/kernel/v4.x/linux-4.10.4.tar.xz
			filename:  "linux-4.10.4.tar.xz",
			hasher:    hasher.NewMD5Hasher(),
			chunksize: (1024 * 1024),
		},
		{
			//https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz
			filename:  "go1.8.linux-amd64.tar.gz",
			hasher:    hasher.NewSHA1Hasher(),
			chunksize: (1024 * 1024 * 1024),
		},
	}

	generateFixtures = flag.Bool("genfixtures", false, "Regenerate fixtures for the test suite.")
)

func init() {
	flag.Parse()

	if *generateFixtures {
		fmt.Printf("*** WARNING: fixtures will be recreated in a few seconds...\n")
		fmt.Printf("Hit Ctrl+C to abort.\n")
		time.Sleep(time.Second * 1)
		fmt.Printf("\n")
		RegenerateFixtures()
	}
}

func RegenerateFixtures() {
	for _, tc := range testcases {
		// we wrap this in a func() to be able to use the defer statement
		func() {
			testfile := filepath.Join("fixtures", tc.filename)

			// open the source file
			var source SourceFile
			source, err := OpenLocalSource(testfile)
			if err != nil {
				panic(fmt.Errorf("Failed to open test file: %s: %s", tc.filename, err.Error()))
			}
			defer source.Close()

			// recreate the chunk database
			err = source.BuildCache(&tc.hasher, tc.chunksize)
			if err != nil {
				panic(fmt.Errorf(err.Error()))
			}

			// re-read all chunks and write them to a file
			fixturesfile := filepath.Join("fixtures", tc.filename+".cachedump")
			f, err := os.OpenFile(fixturesfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				panic(fmt.Errorf("Failed to create new chunk check file: %s: %s", fixturesfile, err.Error()))
			}
			defer f.Close()

			info, err := source.GetFileInfo()
			if err != nil {
				panic(fmt.Errorf("Failed to get file info from cache: %s", err.Error()))
			}
			_, err = f.WriteString(fmt.Sprintf("%#v\n", info))
			if err != nil {
				panic(fmt.Errorf("Failed to write test file: %s", err.Error()))
			}

			numOfChunks, chunkStreamChan := source.GetAllChunks()
			_, err = f.WriteString(fmt.Sprintf("Chunks: %d\n", numOfChunks))
			if err != nil {
				panic(fmt.Errorf("Failed to write test file: %s", err.Error()))
			}
			for chunkStream := range chunkStreamChan {
				_, err = f.WriteString(fmt.Sprintf("%#v\n", chunkStream))
				if err != nil {
					panic(fmt.Errorf("Failed to write test file: %s", err.Error()))
				}
			}

			err = f.Sync()
			if err != nil {
				panic(fmt.Errorf("Failed to write test file (sync): %s", err.Error()))
			}
		}()

	}
}

func TestLocalFileCacheGeneration(t *testing.T) {
	for _, tc := range testcases {
		func() {
			testfile := filepath.Join("fixtures", tc.filename)

			// open the source file
			var source SourceFile
			source, err := OpenLocalSource(testfile)
			if err != nil {
				t.Fatalf("[%s] Failed to open test file: %s: %s", tc.filename, tc.filename, err.Error())
			}
			defer source.Close()

			// recreate the chunk database
			err = source.BuildCache(&tc.hasher, tc.chunksize)
			if err != nil {
				t.Fatalf(err.Error())
			}

			// re-read all chunks and write them to a file
			fixturesfile := filepath.Join("fixtures", tc.filename+".cachedump")
			f, err := ioutil.TempFile("fixtures", fmt.Sprintf("test_tmp_%s_", tc.filename))
			if err != nil {
				t.Fatalf("[%s] Failed to create new chunk check file: %s: %s", tc.filename, f.Name(), err.Error())
				return
			}
			defer f.Close()
			defer os.Remove(f.Name())

			info, err := source.GetFileInfo()
			if err != nil {
				t.Fatalf("[%s] Failed to get file info from cache: %s", tc.filename, err.Error())
				return
			}
			_, err = f.WriteString(fmt.Sprintf("%#v\n", info))
			if err != nil {
				t.Fatalf("[%s] Failed to write test file: %s", tc.filename, err.Error())
				return
			}

			numOfChunks, chunkStreamChan := source.GetAllChunks()
			_, err = f.WriteString(fmt.Sprintf("Chunks: %d\n", numOfChunks))
			if err != nil {
				panic(fmt.Errorf("[%s] Failed to write test file: %s", tc.filename, err.Error()))
			}
			for chunkStream := range chunkStreamChan {
				_, err = f.WriteString(fmt.Sprintf("%#v\n", chunkStream))
				if err != nil {
					t.Fatalf("[%s] Failed to write test file: %s", tc.filename, err.Error())
					return
				}
			}

			err = f.Sync()
			if err != nil {
				t.Fatalf("[%s] Failed to write test file (sync): %s", tc.filename, err.Error())
				return
			}

			h1, err := tc.hasher.HashFile(fixturesfile)
			if err != nil {
				t.Fatalf("[%s] Failed to calculate checksum: %s: %s", tc.filename, fixturesfile, err.Error())
				return
			}
			h2, err := tc.hasher.HashFile(f.Name())
			if err != nil {
				t.Fatalf("[%s] Failed to calculate checksum: %s: %s", tc.filename, f.Name(), err.Error())
				return
			}
			if h1 != h2 {
				t.Fatalf("[%s] Checksum is different, test returns different data (%s  %s)", tc.filename, fixturesfile, f.Name())
				return
			}
			fmt.Printf("[%s] Checksum is OK (%s  %s)\n", tc.filename, fixturesfile, f.Name())
		}()

	}
}

func TestLocalFileCopy(t *testing.T) {
	for _, tc := range testcases {
		sourcefile := filepath.Join("fixtures", tc.filename)
		targetfile := filepath.Join("fixtures", fmt.Sprintf("target_%s", tc.filename))

		// make sure the target file doesn't exist
		if _, err := os.Stat(targetfile); err == nil {
			err := os.Remove(targetfile)
			if err != nil {
				t.Fatalf("[%s] Failed to delete target file: %s: %s", tc.filename, targetfile, err.Error())
			}
		}

		err := CopyLocalToLocal(sourcefile, targetfile, &tc.hasher, tc.chunksize)
		if err != nil {
			t.Fatalf("[%s] Failed to copy file: %s -> %s: %s", tc.filename, sourcefile, targetfile, err.Error())
		}

		// validate the checksums of both files on disk (slow)
		h1, err := tc.hasher.HashFile(sourcefile)
		if err != nil {
			t.Errorf("[%s] Failed to calculate checksum: %s: %s", tc.filename, sourcefile, err.Error())
		}
		h2, err := tc.hasher.HashFile(targetfile)
		if err != nil {
			t.Errorf("[%s] Failed to calculate checksum: %s: %s", tc.filename, targetfile, err.Error())
		}
		if h1 != h2 {
			t.Errorf("[%s] Checksum is different! (%s %s)", tc.filename, sourcefile, targetfile)
		}

	}

}
