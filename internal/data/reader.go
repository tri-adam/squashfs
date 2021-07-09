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
}

func NewReader(rdr io.ReaderAt, offset uint64, sizes []uint32, decomp decompress.Decompressor, frag *Fragment) (out *Reader, err error) {
	out = new(Reader)
	out.curReader, err = GetDataBlockReader(rdr, offset, 0, sizes[0], decomp)
	if err != nil {
		if out.curReader != nil {
			out.curReader.Close()
		}
		return
	}
	out.baseRdr = rdr
	out.decomp = decomp
	out.nextOffset = offset + uint64(sizes[0]&^(1<<24))
	out.sizes = sizes[1:]
	out.frag = frag
	return
}

func NewReaderFromInode(rdr io.ReaderAt, decomp decompress.Decompressor, i *components.Inode, fragTable []components.FragBlockEntry) (*Reader, error) {
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
			}
		}
	case components.ExtFileType:
		offset = i.Data.(components.ExtFile).BlockStart
		sizes = i.Data.(components.ExtFile).BlockSizes
		if i.Data.(components.ExtFile).FragIndex != 0xFFFFFFFF {
			frag = &Fragment{
				entry:  fragTable[i.Data.(components.ExtFile).FragIndex],
				offset: i.Data.(components.ExtFile).FragBlockOffset,
			}
		}
	default:
		return nil, errors.New("given inode isn't file type")
	}
	return NewReader(rdr, offset, sizes, decomp, frag)
}

func (d *Reader) setupNextReader() (err error) {
	if len(d.sizes) == 0 {
		if d.frag == nil {
			return io.EOF
		}
		d.curReader.Close()
		d.curReader, err = d.frag.GetDataReader(d.baseRdr, d.decomp)
		if err != nil && d.curReader != nil {
			d.curReader.Close()
		}
		d.frag = nil
		return
	}
	d.curReader.Close()
	d.curReader, err = GetDataBlockReader(d.baseRdr, d.nextOffset, 0, d.sizes[0], d.decomp)
	if err != nil {
		if d.curReader != nil {
			d.curReader.Close()
		}
		return
	}
	d.nextOffset = d.nextOffset + uint64(d.sizes[0]&^(1<<24))
	d.sizes = d.sizes[1:]
	return
}

func (d *Reader) Read(p []byte) (n int, err error) {
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
