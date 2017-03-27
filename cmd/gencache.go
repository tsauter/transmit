// Copyright © 2017 Thorsten Sauter <tsauter@gmx.net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tsauter/transmit/hasher"
	"github.com/tsauter/transmit/transmitlib"
)

// gencacheCmd represents the gencache command
var (
	gencacheCmd = &cobra.Command{
		Use:   "gencache",
		Short: "(Re-)build a chunk cache for a local file",
		Long: `The gencache command reads a local file chunk by chunk and 
creates a local file with these information. This file will be used later
to avoid recalculating the whole chunk database for this file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if sourcefilename == "" {
				fmt.Printf("Filename is missing.\n")
				os.Exit(1)
			}
			if _, err := os.Stat(sourcefilename); os.IsNotExist(err) {
				fmt.Printf("File does not exist: %s\n", sourcefilename)
				os.Exit(1)
			}

			var ghasher hasher.Hasher
			switch strings.ToLower(hashalgo) {
			case "sha256":
				ghasher = hasher.NewSHA256Hasher()
			case "sha1":
				ghasher = hasher.NewSHA1Hasher()
			case "md5":
				ghasher = hasher.NewMD5Hasher()
			default:
				fmt.Printf("unsupported hash algorithm: %s\n", hashalgo)
				os.Exit(1)
			}

			fmt.Printf("Generating cache database for %s (algorithm %s, chunksize %d Bytes)\n", sourcefilename, ghasher.GetName(), chunksize)

			// open the source file
			var source transmitlib.SourceFile
			source, err := transmitlib.OpenLocalSource(sourcefilename)
			if err != nil {
				fmt.Printf("Failed to open test file: %s: %s", sourcefilename, err.Error())
				os.Exit(1)
			}
			defer source.Close()

			// recreate the chunk database
			fmt.Printf("Building local file cache...\n")
			err = source.BuildCache(&ghasher, chunksize)
			if err != nil {
				fmt.Printf("Failed to build cache database: %s", err.Error())
				os.Exit(1)
			}

		},
	}

	// flag variables
	sourcefilename string
	hashalgo       string
	chunksize      int

	force bool
)

func init() {
	RootCmd.AddCommand(gencacheCmd)

	gencacheCmd.PersistentFlags().StringVar(&sourcefilename, "filename", "", "source file for chunk calculation")
	gencacheCmd.PersistentFlags().IntVar(&chunksize, "chunksize", 1024*1024, "size for the individual chunks")
	gencacheCmd.PersistentFlags().StringVar(&hashalgo, "hash-algorithm", "sha1", "which algorithm should be used for calculating the chunks")
	gencacheCmd.PersistentFlags().BoolVar(&force, "force", false, "always overwrite existing cache files")
}
