package data

import (
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

func GetDataBlockReader(r io.ReaderAt, offset uint64, blockOffset, size uint32, decomp decompress.Decompressor) (io.ReadCloser, error) {
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
	_, err = rdr.Read(make([]byte, blockOffset))
	if err != nil {
		rdr.Close()
		return nil, err
	}
	return rdr, nil
}
