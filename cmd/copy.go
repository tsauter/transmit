// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

// copyCmd represents the copy command
var (
	copyCmd = &cobra.Command{
		Use:   "copy",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			// make sure the two required parameters source and target are specified
			if (sourcefilename == "") || (targetfilename == "") {
				fmt.Printf("Missing source or target file.\n")
				os.Exit(1)
			}

			if strings.HasPrefix(targetfilename, "http://") {
				fmt.Printf("Target file can not be a remote file (http)\n")
				os.Exit(1)
			}

			// load the haser based on the user settings
			var ghasher hasher.Hasher
			switch hashalgo {
			case "sha256":
				ghasher = hasher.NewSHA256Hasher()
			case "sha1":
				ghasher = hasher.NewSHA1Hasher()
			case "md5":
				ghasher = hasher.NewMD5Hasher()
			default:
				fmt.Printf("Unsupported hash method: %s", hashalgo)
				os.Exit(1)
			}

			fmt.Printf("Copy file %s to %s (algorithm %s, chunksize %d Bytes)\n", sourcefilename, targetfilename, ghasher.GetName(), chunksize)

			var err error

			if strings.HasPrefix(sourcefilename, "http://") {
				err = transmitlib.CopyHttpToLocal(sourcefilename, targetfilename, &ghasher, chunksize)

			} else {
				if _, err := os.Stat(sourcefilename); os.IsNotExist(err) {
					fmt.Printf("File does not exist: %s\n", sourcefilename)
					os.Exit(1)
				}

				err = transmitlib.CopyLocalToLocal(sourcefilename, targetfilename, &ghasher, chunksize)
			}

			if err != nil {
				fmt.Printf("Failed to copy file: %s -> %s: %s", sourcefilename, targetfilename, err.Error())
				os.Exit(1)
			}
			fmt.Printf("File successfully copied!\n")

		},
	}

	// flag variables
	//sourcefilename string
	targetfilename string
	//hashalgo       string
	//chunksize      int
)

func init() {
	RootCmd.AddCommand(copyCmd)

	copyCmd.PersistentFlags().StringVar(&sourcefilename, "sourcefile", "", "source file for copying")
	copyCmd.PersistentFlags().StringVar(&targetfilename, "targetfile", "", "target file for copying")
	copyCmd.PersistentFlags().IntVar(&chunksize, "chunksize", 1024*1024, "size for the individual chunks")
	copyCmd.PersistentFlags().StringVar(&hashalgo, "hash-algorithm", "sha1", "which algorithm should be used for calculating the chunks")
}
