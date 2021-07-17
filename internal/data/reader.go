package data

import (
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/decompress"
)

type Reader struct {
	curReader  io.ReadCloser
	baseRdr    io.ReaderAt
	decomp     decompress.Decompressor
	frag       *Fragment
	sizes      []uint32
	nextOffset uint64
	blockSize  uint32
}

func NewReader(rdr io.ReaderAt, offset uint64, sizes []uint32, blockSize uint32, decomp decompress.Decompressor, frag *Fragment) (out *Reader, err error) {
	out = new(Reader)
	if len(sizes) == 0 {
		if frag == nil {
			return
		}
		out.curReader, err = frag.GetDataReader(rdr, decomp)
		return
	}
	out.baseRdr = rdr
	out.decomp = decomp
	out.frag = frag
	out.sizes = sizes
	out.nextOffset = offset
	out.blockSize = blockSize
	return
}

func NewReaderFromInode(rdr io.ReaderAt, blockSize uint32, decomp decompress.Decompressor, i *components.Inode, fragTable []components.FragBlockEntry) (*Reader, error) {
	var offset uint64
	var sizes []uint32
	var frag *Fragment
	switch i.Type {
	case components.FileType:
		offset = uint64(i.Data.(components.File).BlockStart)
		sizes = i.Data.(components.File).BlockSizes
		if i.Data.(components.File).FragIndex != 0xFFFFFFFF {
			frag = &Fragment{
				entry:  fragTable[i.Data.(components.File).FragIndex],
				offset: i.Data.(components.File).FragBlockOffset,
				size:   i.Data.(components.File).Size % blockSize,
			}
		}
	case components.ExtFileType:
		offset = i.Data.(components.ExtFile).BlockStart
		sizes = i.Data.(components.ExtFile).BlockSizes
		if i.Data.(components.ExtFile).FragIndex != 0xFFFFFFFF {
			frag = &Fragment{
				entry:  fragTable[i.Data.(components.ExtFile).FragIndex],
				offset: i.Data.(components.ExtFile).FragBlockOffset,
				size:   uint32(i.Data.(components.ExtFile).Size % uint64(blockSize)),
			}
		}
	default:
		return nil, errors.New("given inode isn't file type")
	}
	return NewReader(rdr, offset, sizes, blockSize, decomp, frag)
}

func (d *Reader) setupNextReader() (err error) {
	if len(d.sizes) == 0 {
		if d.frag == nil {
			return io.EOF
		}
		if d.curReader != nil {
			d.curReader.Close()
		}
		d.curReader, err = d.frag.GetDataReader(d.baseRdr, d.decomp)
		if err != nil && d.curReader != nil {
			d.curReader.Close()
		}
		d.frag = nil
		return
	}
	if d.curReader != nil {
		d.curReader.Close()
	}
	if d.sizes[0] == 0 {
		d.curReader = &zeroReader{
			size: d.blockSize,
		}
	} else {
		d.curReader, err = GetDataBlockReader(d.baseRdr, d.nextOffset, 0, d.sizes[0], d.decomp, 0)
		if err != nil {
			if d.curReader != nil {
				d.curReader.Close()
			}
			return
		}
		d.nextOffset = d.nextOffset + uint64(d.sizes[0]&^(1<<24))
	}
	d.sizes = d.sizes[1:]
	return
}

func (d *Reader) Read(p []byte) (n int, err error) {
	if d.curReader == nil {
		err = d.setupNextReader()
		if err != nil {
			return
		}
	}
	n, err = d.curReader.Read(p)
	if err == nil {
		return
	}
	if err != io.EOF && err != nil {
		return
	}
	err = d.setupNextReader()
	if err != nil {
		return
	}
	tmp := make([]byte, len(p)-n)
	add, err := d.Read(tmp)
	for i := range p[n : n+add] {
		p[n+i] = tmp[i]
	}
	n += add
	return
}

func (d *Reader) Close() error {
	if d.baseRdr != nil {
		return d.curReader.Close()
	}
	return nil
}

func (d *Reader) WriteTo(w io.Writer) (n int64, err error) {
	if len(d.sizes) == 0 && d.frag == nil {
		return
	}
	outChan := make(chan *writerToReturn)
	offset := d.nextOffset
	blocks := len(d.sizes)
	if d.frag != nil {
		blocks++
	}
	for i, s := range d.sizes {
		go func(size uint32, offset uint64, i int) {
			out := new(writerToReturn)
			out.index = i
			var rdr io.ReadCloser
			rdr, out.err = GetDataBlockReader(d.baseRdr, offset, 0, size, d.decomp, 0)
			if out.err != nil {
				if rdr != nil {
					rdr.Close()
				}
				outChan <- out
				return
			}
			defer rdr.Close()
			out.data, out.err = io.ReadAll(rdr)
			outChan <- out
		}(s, offset, i)
		offset += uint64(s &^ (1 << 24))
	}
	if d.frag != nil {
		go func() {
			out := new(writerToReturn)
			out.index = blocks - 1
			var rdr io.ReadCloser
			rdr, err = d.frag.GetDataReader(d.baseRdr, d.decomp)
			if out.err != nil {
				if rdr != nil {
					rdr.Close()
				}
				outChan <- out
				return
			}
			defer rdr.Close()
			out.data, out.err = io.ReadAll(rdr)
			outChan <- out
		}()
	}
	curBlock := 0
	tmp := 0
	var backLog []*writerToReturn
mainLoop:
	for curBlock < blocks {
		if len(backLog) > 0 {
			for _, b := range backLog {
				if b.index == curBlock {
					tmp, err = w.Write(b.data)
					n += int64(tmp)
					if err != nil {
						return
					}
					curBlock++
					continue mainLoop
				}
			}
		}
		b := <-outChan
		if b.err != nil {
			return n, b.err
		}
		if b.index == curBlock {
			tmp, err = w.Write(b.data)
			n += int64(tmp)
			if err != nil {
				return
			}
			curBlock++
			continue
		} else {
			backLog = append(backLog, b)
		}
	}
	return
}

type writerToReturn struct {
	err   error
	data  []byte
	index int
}
