package squashfs

import (
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
