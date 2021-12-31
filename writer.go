package squashfs

import (
	"io/fs"
	"math"

	"github.com/CalebQ42/squashfs/internal/compression"
)

//Writer allows for creating squashfs archives.
type Writer struct {
	compressor compression.Compressor
	super      superblock
}

func NewWriter() *Writer {
	return NewWriterWithOptions(DefaultWriterOptions)
}

//NewWriterWithOptions creates a new squashfs.Writer with the given options. If invalid options are given, defaults are used.
func NewWriterWithOptions(options WriterOptions) *Writer {
	var out Writer
	if options.BlockSize < 4096 || options.BlockSize > 1048576 {
		out.super.BlockSize = 1048576
	} else {
		out.super.BlockSize = options.BlockSize
	}
	switch options.Compression {
	case GzipCompression:
		out.compressor = &compression.Gzip{}
	case LzmaCompression:
		out.compressor = &compression.Lzma{}
	case XzCompression:
		out.compressor = &compression.Xz{}
	case Lz4Compression:
		out.compressor = &compression.Lz4{}
	default:
		out.compressor = &compression.Zstd{}
	}
	out.super.Flags = options.SuperblockFlags.ToUint()
	out.super.BlockLog = uint16(math.Log2(float64(out.super.BlockSize)))
	out.super.Magic = magic
	return &out
}

//AddFile adds an fs.File to the root of the archive.
//NOTE: Regular files will not be read until Write is called.
func (w *Writer) AddFile(file fs.File) (err error) {
	return nil
}

//AddFileAt adds an fs.File to the path given.
//If a folder was not previously added, it's added with 0755 permissions and no particular owner or group.
//NOTE: Regular files will not be read until Write is called.
func (w *Writer) AddFileAt(file fs.File, path string) (err error) {
	return nil
}
