package squashfs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/metadata"
)

func (r Reader) parseInodeRef(inodeRef uint64) (*components.Inode, error) {
	return r.parseInode(inodeRef>>16, inodeRef&^0xFFFFFFFF0000)
}

func (r Reader) parseInode(offset, blockOffset uint64) (*components.Inode, error) {
	var data []byte
	block, nextBlock, err := metadata.ReadBlockAt(r.rdr, int64(offset), r.decomp)
	if err != nil {
		return nil, err
	}
	data = block[blockOffset:]
	inode := new(components.Inode)
	_ = inode
	if len(data) < binary.Size(inode.InodeHeader) {
		block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
		if err != nil {
			return nil, err
		}
		data = append(data, block...)
	}
	err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &inode.InodeHeader)
	if err != nil {
		return nil, err
	}
	data = data[components.InodeHeaderSize:]
	switch inode.Type {
	case components.DirType:
		inode.Data = components.Dir{}
	case components.FileType:
		inode.Data = components.File{}
	case components.SymType:
		inode.Data = components.Sym{}
	case components.BlockType:
		fallthrough
	case components.CharType:
		inode.Data = components.Device{}
	case components.FifoType:
		fallthrough
	case components.SocketType:
		inode.Data = components.IPC{}
	case components.ExtDirType:
		inode.Data = components.ExtDir{}
	case components.ExtFileType:
		inode.Data = components.ExtFile{}
	case components.ExtSymType:
		inode.Data = components.ExtSym{}
	case components.ExtBlockType:
		fallthrough
	case components.ExtCharType:
		inode.Data = components.ExtDevice{}
	case components.ExtFifoType:
		fallthrough
	case components.ExtSocketType:
		inode.Data = components.ExtIPC{}
	default:
		return nil, errors.New("inode type is unsupported: " + strconv.Itoa(int(inode.Type)))
	}
	if d, is := inode.Data.(components.Dir); is {
		_ = d
	} else {
		for len(data) < binary.Size(inode.Data) {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &inode.Data)
		if err != nil {
			return nil, err
		}
	}
	return inode, nil
}
