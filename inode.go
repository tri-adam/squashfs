package squashfs

import (
	"github.com/CalebQ42/squashfs/internal/components"
)

func (r Reader) parseInode(offset, blockOffset uint64) (*components.Inode, error) {
	//TODO
	//read metadata block
	inode := new(components.Inode)
	_ = inode
	//check if header is all there and possibly read another block
	//Read inode type and read particular inode type, reading new blocks as necessary
	return nil, nil
}

func (r Reader) parseInodeRef(inodeRef uint64) (*components.Inode, error) {
	return r.parseInode(inodeRef>>16, inodeRef&^0xFFFFFFFF0000)
}
