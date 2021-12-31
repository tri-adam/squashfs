package squashfs

import "time"

//WriterOptions dictates some of the options on how a squashfs archive is made.
//BlockSize must be between 4096 and 1048576. If not within the range, the default value (1048576) is used.
//NoXattr is, for right now, is always true since this library doesn't currently support xattr.
//If ModTime.isZero(), the current time when the archive is writen is used.
//
//NOTE: Lzo Compression is NOT supported for writing.
type WriterOptions struct {
	ModTime     *time.Time
	BlockSize   uint32
	Compression CompressionType
	SuperblockFlags
}

//DeafultWriterOptions provides a starting point for setting WriterOptions.
//Note: Zstd compression is chosen by default due to it's benefits vs gzip (squashfs' official default).
var DefaultWriterOptions = WriterOptions{
	BlockSize:   1048576, //1Mb blocks
	Compression: ZstdCompression,
}
