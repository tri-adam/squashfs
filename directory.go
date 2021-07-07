package squashfs

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/metadata"
)

func (r Reader) getDirEntriesFromInode(i *components.Inode) (out []components.DirEntry, err error) {
	var offset, blockOffset, size int64
	switch i.Type {
	case components.DirType:
		offset = int64(i.Data.(components.Dir).DirIndex)
		blockOffset = int64(i.Data.(components.Dir).BlockOffset)
		size = int64(i.Data.(components.Dir).FileSize)
	case components.ExtDirType:
		offset = int64(i.Data.(components.ExtDir).DirIndex)
		blockOffset = int64(i.Data.(components.ExtDir).BlockOffset)
		size = int64(i.Data.(components.ExtDir).FileSize)
	default:
		return nil, errors.New("given inode isn't a dir type")
	}
	offset += int64(r.super.DirTableStart)
	hdr := make([]components.DirHeader, 1)
	var data []byte
	for int64(len(data)) < 12+blockOffset {
		data, offset, err = metadata.ReadBlockAt(r.rdr, offset, r.decomp)
		if err != nil {
			return
		}
	}
	data = data[blockOffset:]
	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &hdr[0])
	if err != nil {
		return
	}
	data = data[12:]

}
