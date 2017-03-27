[![GPL Licence](https://badges.frapsoft.com/os/gpl/gpl.svg?v=103)](https://opensource.org/licenses/GPL-3.0/)

# transmit
Robust and fault tolerant file transfer system. 

Transmit copies a local or remote file. The copy process can be interrupted at any time, and resumed without transfering the complete file again. 

The file is divided into small pieces, and a checksum is computed for each of these pieces. All checksums are written to a database next to the source file. During the copying process, all chunks are compared and only when the checksum of the chunk is different, the real data for that chunk is transferred.

During the copy process the following steps are performed:
1. Open the chunk database of the source file
2. Build a temporary chunk database of the local file, if exists
3. Compare each chunk of source and target file
4. Copy data of the source chunk to the target chunk, if checksum is not equal
5. Compare the checksum of the complete copied target file with the source
6. Delete target cache database

If the file will be transfered over the network, the chunk database for the source file must be created on the source computer. Otherwise all data for caculating the checksums will be transfered over the network!

Hint: MD5 and SHA1 are weak. Please consider the use of SHA256 instead.


## Usage

### Generating source chunk database

The following command generates the chunk database for the source file. The database does not need any updates as long as the source file has not been modified:

```
transfer gencache --filename=bigsourcefile.zip
```

The name of the cache file will be ```bigsourcefile.zip.tcache.db```.

### Copy the file

To copy the file, the following command can be used. 

```
transfer copy --sourcefile=bigsourcefile.zip --targetfile=bigtarget.zip
```

The cache file for the target file will be removed automatically.

## Advanced usage

The following optional parameters exist:

* --chunksize: use different length for each chunk (the chunksize must match source and target chunk database)
* --hash-algorithm: which algorithm is used for the checksums (md5, sha1, sha256 (must be equal between source and target database)
