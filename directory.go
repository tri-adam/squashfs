package squashfs

import (
	"encoding/binary"
	"errors"
	"io"
	"io/fs"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/metadata"
)

type dirEntry struct {
	r *Reader

	ent   components.DirEntry
	Start uint32
}

func (r Reader) getDirEntriesFromInode(i *components.Inode) (out []dirEntry, err error) {
	var offset, blockOffset, size uint64
	switch i.Type {
	case components.DirType:
		offset = uint64(i.Data.(components.Dir).DirIndex)
		blockOffset = uint64(i.Data.(components.Dir).BlockOffset)
		size = uint64(i.Data.(components.Dir).FileSize)
	case components.ExtDirType:
		offset = uint64(i.Data.(components.ExtDir).DirIndex)
		blockOffset = uint64(i.Data.(components.ExtDir).BlockOffset)
		size = uint64(i.Data.(components.ExtDir).FileSize)
	default:
		return nil, errors.New("given inode isn't a dir type")
	}
	offset += uint64(r.super.DirTableStart)
	metRdr, err := metadata.NewReader(r.rdr, offset, blockOffset, r.decomp)
	if err != nil {
		return
	}
	curSize := 0
hdrLoop:
	for curSize+12 < int(size) {
		var hdr components.DirHeader
		err = binary.Read(metRdr, binary.LittleEndian, &hdr)
		if err != nil {
			return
		}
		curSize += 12
		for i := 0; i < int(hdr.Count)+1; i++ {
			var tmp components.DirEntry
			err = binary.Read(metRdr, binary.LittleEndian, &tmp.DirEntryBase)
			if err == io.EOF {
				continue hdrLoop
			}
			if err != nil {
				return
			}
			tmp.Name = make([]byte, tmp.NameSize+1)
			err = binary.Read(metRdr, binary.LittleEndian, &tmp.Name)
			if err != nil {
				return
			}
			curSize += 8 + 1 + int(tmp.NameSize)
			out = append(out, dirEntry{
				ent:   tmp,
				Start: hdr.Start,
				r:     &r,
			})
		}
	}
	return
}

func (d dirEntry) Info() (fs.FileInfo, error) {
	i, err := d.r.dirEntryToInode(d)
	if err != nil {
		return nil, err
	}
	return FileInfo{
		i:   i,
		r:   d.r,
		ent: d,
	}, nil
}

func (d dirEntry) IsDir() bool {
	return d.ent.Type == components.DirType
}

func (d dirEntry) Name() string {
	return string(d.ent.Name)
}

func (d dirEntry) Type() fs.FileMode {
	switch d.ent.Type {
	case components.DirType:
		return fs.ModeDir
	case components.SymType:
		return fs.ModeSymlink
	case components.BlockType:
		fallthrough
	case components.CharType:
		return fs.ModeType
	case components.FileType:
		return 0
	default:
		return fs.ModeIrregular
	}
}
