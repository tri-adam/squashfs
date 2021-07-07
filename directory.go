package squashfs

import (
	"encoding/binary"
	"errors"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/metadata"
)

func (r Reader) getDirEntriesFromInode(i *components.Inode) (out []components.DirEntry, err error) {
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
	hdr := make([]components.DirHeader, 0)
	out = make([]components.DirEntry, 0)
	for binary.Size(hdr)+binary.Size(out) < int(size) {
		hdr = append(hdr, components.DirHeader{})
		err = binary.Read(metRdr, binary.LittleEndian, &hdr[len(hdr)-1])
		if err != nil {
			return
		}
		outTmp := make([]components.DirEntry, hdr[len(hdr)-1].Count)
		for i := range outTmp {
			err = binary.Read(metRdr, binary.LittleEndian, &outTmp[i].DirEntryBase)
			if err != nil {
				return
			}
			outTmp[i].Name = make([]byte, outTmp[i].NameSize)
			err = binary.Read(metRdr, binary.LittleEndian, &outTmp[i].Name)
			if err != nil {
				return
			}
		}
	}
	return
}
