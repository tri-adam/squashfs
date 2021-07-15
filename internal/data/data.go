package data

import (
	"fmt"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

func GetDataBlockReader(r io.ReaderAt, offset uint64, blockOffset, size uint32, decomp decompress.Decompressor, limit uint32) (io.ReadCloser, error) {
	comp := size&(1<<24) != (1 << 24)
	size &^= (1 << 24)
	if !comp {
		return io.NopCloser(io.NewSectionReader(r, int64(offset+uint64(blockOffset)), int64(size-blockOffset))), nil
	}
	secRdr := io.NewSectionReader(r, int64(offset), int64(size))
	rdr, err := decomp.Reader(secRdr)
	if err != nil {
		if rdr != nil {
			rdr.Close()
		}
		return nil, err
	}
	tmp := make([]byte, blockOffset)
	n, err := rdr.Read(tmp)
	fmt.Println("HIIIIEI", n, blockOffset)
	if err != nil {
		rdr.Close()
		return nil, err
	}
	if limit != 0 {
		return NewLimitReaderCloser(rdr, int64(limit)), nil
	}
	return rdr, nil
}

type LimitReaderCloser struct {
	rdr   io.ReadCloser
	limit io.Reader
}

func NewLimitReaderCloser(r io.ReadCloser, n int64) LimitReaderCloser {
	return LimitReaderCloser{
		rdr:   r,
		limit: io.LimitReader(r, n),
	}
}

func (l LimitReaderCloser) Read(p []byte) (n int, err error) {
	n, err = l.limit.Read(p)
	if err == io.EOF {
		l.rdr.Close()
	}
	return
}

func (l LimitReaderCloser) Close() error {
	return l.rdr.Close()
}
