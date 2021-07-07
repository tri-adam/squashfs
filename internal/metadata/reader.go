package metadata

import (
	"encoding/binary"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

type Reader struct {
	rdr    io.ReaderAt
	decomp decompress.Decompressor
	data   []byte

	nextBlock uint64
}

func NewReader(rdr io.ReaderAt, offset, startOffset uint64, decomp decompress.Decompressor) (r *Reader, err error) {
	r = new(Reader)
	r.nextBlock = offset
	r.decomp = decomp
	r.rdr = rdr
	for int(startOffset) > len(r.data) {
		err = r.readNewBlock()
		if err != nil {
			return
		}
	}
	r.data = r.data[startOffset:]
	return
}

func (r *Reader) readNewBlock() (err error) {
	var hdr uint16
	sec := io.NewSectionReader(r.rdr, int64(r.nextBlock), 2)
	err = binary.Read(sec, binary.LittleEndian, &hdr)
	if err != nil {
		return
	}
	newData := make([]byte, hdr&0x7FFF)
	_, err = r.rdr.ReadAt(newData, int64(r.nextBlock)+2)
	if err != nil {
		return
	}
	if hdr&0x8000 != 0x8000 {
		newData, err = decompress.Decompress(r.decomp, newData)
		if err != nil {
			return
		}
	}
	r.data = append(r.data, newData...)
	r.nextBlock += 2 + uint64(hdr&0x7FFF)
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	for len(r.data) < len(p) {
		err = r.readNewBlock()
		if err != nil {
			return
		}
	}
	for i := range p {
		p[i] = r.data[i]
	}
	r.data = r.data[len(p):]
	n = len(p)
	return
}

func (r *Reader) Reset(offset, startOffset uint64) (err error) {
	r.nextBlock = offset
	err = r.readNewBlock()
	if err != nil {
		return
	}
	r.data = r.data[startOffset:]
	return
}
