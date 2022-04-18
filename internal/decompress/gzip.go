package decompress

import (
	"io"

	"github.com/CalebQ42/squashfs/internal/readerattoreader"
	"github.com/klauspost/compress/gzip"
)

type GZip struct{}

func (g GZip) ReaderAt(r io.ReaderAt, offset int64) (io.Reader, error) {
	return gzip.NewReader(readerattoreader.NewReader(r, offset))
}

func (g GZip) Reader(src io.Reader) (io.Reader, error) {
	return gzip.NewReader(src)
}
