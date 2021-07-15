package squashfs

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/metadata"
)

type dirEntry struct {
	components.DirEntry
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
				DirEntry: tmp,
				Start:    hdr.Start,
			})
		}
	}
	return
}
