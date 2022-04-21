package squashfs

import (
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/data"
	"github.com/CalebQ42/squashfs/internal/directory"
	"github.com/CalebQ42/squashfs/internal/inode"
	"github.com/CalebQ42/squashfs/internal/metadata"
	"github.com/CalebQ42/squashfs/internal/readerattoreader"
)

func (r Reader) inodeFromRef(ref uint64) (i inode.Inode, err error) {
	offset, meta := (ref>>16)+r.s.inodeTableStart, ref&0xFFFF
	rdr, err := metadata.NewReader(readerattoreader.NewReader(r.r, int64(offset)), r.d)
	if err != nil {
		return
	}
	_, err = rdr.Read(make([]byte, meta))
	if err != nil {
		return
	}
	return inode.Read(rdr, r.s.blockSize)
}

func (r Reader) getData(i inode.Inode) (io.Reader, error) {
	var offset uint64
	var blockOffset uint32
	var blockSizes []uint32
	if i.Type == inode.Fil {
		offset = uint64(i.Data.(inode.File).Offset)
		blockOffset = i.Data.(inode.File).BlockStart
		blockSizes = i.Data.(inode.File).BlockSizes
	} else if i.Type == inode.EFil {
		offset = uint64(i.Data.(inode.EFile).Offset)
		blockOffset = i.Data.(inode.EFile).BlockStart
		blockSizes = i.Data.(inode.EFile).BlockSizes
	} else {
		return nil, errors.New("getData called on non-file type")
	}
	offset += r.s.inodeTableStart
	rdr, err := data.NewReader(readerattoreader.NewReader(r.r, int64(offset)), r.d, blockSizes)
	if err != nil {
		return nil, err
	}
	_, err = rdr.Read(make([]byte, blockOffset))
	if err != nil {
		return nil, err
	}
	return rdr, nil
}

func (r Reader) readDirectory(i inode.Inode) ([]directory.Entry, error) {
	var offset uint64
	var blockOffset uint16
	var size uint32
	if i.Type == inode.Dir {
		offset = uint64(i.Data.(inode.Directory).BlockStart)
		blockOffset = i.Data.(inode.Directory).Offset
		size = uint32(i.Data.(inode.Directory).Size)
	} else if i.Type == inode.EDir {
		offset = uint64(i.Data.(inode.EDirectory).BlockStart)
		blockOffset = i.Data.(inode.EDirectory).Offset
		size = i.Data.(inode.EDirectory).Size
	} else {
		return nil, errors.New("readDirectory called on non-directory type")
	}
	rdr, err := metadata.NewReader(readerattoreader.NewReader(r.r, int64(offset)), r.d)
	if err != nil {
		return nil, err
	}
	_, err = rdr.Read(make([]byte, blockOffset))
	if err != nil {
		return nil, err
	}
	return directory.ReadEntries(rdr, size)
}
