package squashfs

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
	"github.com/CalebQ42/squashfs/internal/readerattoreader"
)

type Reader struct {
	d decompress.Decompressor
	r io.ReaderAt
	s superblock
}

var (
	ErrorMagic = errors.New("magic incorrect. probably not reading squashfs archive")
)

const (
	GZipCompression = uint16(iota + 1)
	LZMACompression
	LZOCompression
	XZCompression
	LZ4Compression
	ZSTDCompression
)

func NewReader(r io.ReaderAt) (*Reader, error) {
	var squash Reader
	squash.r = r
	err := binary.Read(readerattoreader.NewReader(r, 0), binary.LittleEndian, &squash.s)
	if err != nil {
		return nil, err
	}
	if !squash.s.hasMagic() {
		return nil, ErrorMagic
	}
	switch squash.s.compType {
	case GZipCompression:
		squash.d = decompress.GZip{}
	case ZSTDCompression:
		squash.d = decompress.Zstd{}
	default:
		return nil, errors.New("uh, I need to do this, OR something if very wrong")
	}

	//TODO:
	//	FragOffsets
	//	IDTable
	//	Parse Root Inode
	return &squash, nil
}
