package data

import (
	"bytes"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

type Reader struct {
	r          io.Reader
	dRdr       io.Reader
	d          decompress.Decompressor
	blockSizes []uint32
	left       uint32
	//advance on next Read call
	adv bool
}

func NewReader(r io.Reader, d decompress.Decompressor, blockSizes []uint32) (*Reader, error) {
	var out Reader
	out.d = d
	out.r = r
	out.blockSizes = blockSizes
	err := out.advance()
	return &out, err
}

func (r *Reader) advance() (err error) {
	if len(r.blockSizes) == 0 {
		return io.EOF
	}
	size := r.blockSizes[0]
	r.blockSizes = r.blockSizes[1:]
	comp := size&(1<<24) != (1 << 24)
	size = size &^ (1 << 24)
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

type readOut struct {
	d          decompress.Decompressor
	err        error
	unCompData []byte
	outData    []byte
	index      int
}

func (r *readOut) decomp(out chan readOut) {
	var rdr io.Reader
	rdr, r.err = r.d.Reader(bytes.NewReader(r.unCompData))
	if r.err != nil {
		out <- *r
		return
	}
	r.outData, r.err = io.ReadAll(rdr)
	out <- *r
}

func (r *Reader) WriteTo(w io.Writer) (n int64, err error) {
	out := make(chan readOut, len(r.blockSizes))
	for i := range r.blockSizes {
		size := r.blockSizes[i]
		comp := size&(1<<24) != (1 << 24)
		size = size &^ (1 << 24)
		tmp := make([]byte, size)
		_, err = r.r.Read(tmp)
		if err != nil {
			return
		}
		if !comp {
			out <- readOut{
				outData: tmp,
				index:   i,
			}
		} else {
			o := &readOut{
				d:          r.d,
				unCompData: tmp,
				index:      i,
			}
			go o.decomp(out)
		}
	}
	var cache []readOut
	curInd, tmpN := 0, 0
	for rd, ok := <-out; ok; {
		if rd.err != nil {
			err = rd.err
			return
		}
		if curInd != rd.index {
			cache = append(cache, rd)
			continue
		}
		tmpN, err = w.Write(rd.outData)
		n += int64(tmpN)
		if err != nil {
			return
		}
		curInd++
		for i := range cache {
			if cache[i].index == curInd {
				tmpN, err = w.Write(rd.outData)
				n += int64(tmpN)
				if err != nil {
					return
				}
				curInd++
			}
		}
	}
	return
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
