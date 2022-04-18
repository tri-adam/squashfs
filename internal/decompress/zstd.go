package decompress

import (
	"io"

	"github.com/CalebQ42/squashfs/internal/readerattoreader"
	"github.com/klauspost/compress/zstd"
)

type Zstd struct{}

func (z Zstd) ReaderAt(r io.ReaderAt, offset int64) (io.Reader, error) {
	return zstd.NewReader(readerattoreader.NewReader(r, offset))
}

func (z Zstd) Reader(src io.Reader) (io.Reader, error) {
	return zstd.NewReader(src)
}
