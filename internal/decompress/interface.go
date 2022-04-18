package decompress

import "io"

type Decompressor interface {
	ReaderAt(src io.ReaderAt, offset int64) (io.Reader, error)
	Reader(src io.Reader) (io.Reader, error)
}
