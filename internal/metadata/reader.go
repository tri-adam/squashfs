package metadata

import (
	"encoding/binary"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
	"github.com/CalebQ42/squashfs/internal/readerattoreader"
)

type Reader struct {
	r    io.Reader
	dRdr io.Reader
	d    decompress.Decompressor
	//left to read before next metadata block.
	//If 0, current block is compressed.
	left uint16
	//advance on next Read call
	adv bool
}

func NewReader(ra io.ReaderAt, start uint64, d decompress.Decompressor) (*Reader, error) {
	var out Reader
	out.d = d
	out.r = readerattoreader.NewReader(ra, int64(start))
	err := out.advance()
	return &out, err
}

func (r *Reader) advance() error {
	var size uint16
	err := binary.Read(r.r, binary.LittleEndian, &size)
	if err != nil {
		return err
	}
	comp := size&0x8000 != 0x8000
	size = size & 0x7FFF
	if !comp {
		r.left = size
	} else {
		r.dRdr, err = r.d.Reader(r.r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.adv {
		err = r.advance()
		if err != nil {
			return
		}
	}
	if r.left == 0 {
		n, err = r.dRdr.Read(p)
		if err == io.EOF {
			err = r.advance()
			if err != nil {
				return n, err
			}
			tmp := make([]byte, len(p)-n)
			var tmpN int
			tmpN, err = r.Read(tmp)
			for i := range tmp {
				p[n+i] = tmp[i]
			}
			n += tmpN
			if n == len(p) && err == io.EOF {
				r.adv = true
				err = nil
			}
		}
		return
	}
	if len(p) < int(r.left) {
		return r.r.Read(p)
	}
	if len(p) > int(r.left) {
		tmp := make([]byte, r.left)
		n, err = r.r.Read(tmp)
		copy(p, tmp)
		if err != nil {
			return
		}
		err = r.advance()
		if err != nil {
			return
		}
		tmp = make([]byte, len(p)-int(r.left))
		var tmpN int
		tmpN, err = r.Read(tmp)
		for i := range tmp {
			p[n+i] = tmp[i]
		}
		n += tmpN
	} else if len(p) == int(r.left) {
		n, err = r.r.Read(p)
		if n != len(p) {
			return
		}
		r.adv = true
	}
	return
}
