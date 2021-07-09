package squashfs

import (
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/metadata"
)

func (r Reader) dirEntryToInode(ent *dirEntry) (*components.Inode, error) {
	return r.parseInode(uint64(ent.Start), uint64(ent.Offset))
}

func (r Reader) parseInodeRef(inodeRef uint64) (*components.Inode, error) {
	return r.parseInode(inodeRef>>16, inodeRef&^0xFFFFFFFF0000)
}

func (r Reader) parseInode(offset, blockOffset uint64) (*components.Inode, error) {
	inode := new(components.Inode)
	metRdr, err := metadata.NewReader(r.rdr, offset, blockOffset, r.decomp)
	if err != nil {
		return nil, err
	}
	err = binary.Read(metRdr, binary.LittleEndian, &inode.InodeHeader)
	if err != nil {
		return nil, err
	}
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
	switch inode.Type {
	case components.ExtDirType:
		d := inode.Data.(components.ExtDir)
		err = binary.Read(metRdr, binary.LittleEndian, &d.ExtDirBase)
		if err != nil {
			return nil, err
		}
		d.Indexes = make([]components.DirIndex, d.IndexCount)
		for i := range d.Indexes {
			err = binary.Read(metRdr, binary.LittleEndian, &d.Indexes[i].DirIndexBase)
			if err != nil {
				return nil, err
			}
			d.Indexes[i].Name = make([]byte, d.Indexes[i].NameSize)
			err = binary.Read(metRdr, binary.LittleEndian, &d.Indexes[i].Name)
			if err != nil {
				return nil, err
			}
		}
	case components.FileType:
		f := inode.Data.(components.File)
		err = binary.Read(metRdr, binary.LittleEndian, &f.FileBase)
		if err != nil {
			return nil, err
		}
		sizeNum := f.Size / r.super.BlockSize
		if f.Size&r.super.BlockSize > 0 {
			sizeNum++
		}
		f.BlockSizes = make([]uint32, sizeNum)
		err = binary.Read(metRdr, binary.LittleEndian, &f.BlockSizes)
		if err != nil {
			return nil, err
		}
	case components.ExtFileType:
		f := inode.Data.(components.ExtFile)
		err = binary.Read(metRdr, binary.LittleEndian, &f.ExtFileBase)
		if err != nil {
			return nil, err
		}
		sizeNum := f.Size / uint64(r.super.BlockSize)
		if f.Size&uint64(r.super.BlockSize) > 0 {
			sizeNum++
		}
		f.BlockSizes = make([]uint32, sizeNum)
		err = binary.Read(metRdr, binary.LittleEndian, &f.BlockSizes)
		if err != nil {
			return nil, err
		}
	case components.SymType:
		s := inode.Data.(components.Sym)
		err = binary.Read(metRdr, binary.LittleEndian, &s.SymBase)
		if err != nil {
			return nil, err
		}
		s.Path = make([]byte, s.PathSize)
		err = binary.Read(metRdr, binary.LittleEndian, &s.Path)
		if err != nil {
			return nil, err
		}
	case components.ExtSymType:
		s := inode.Data.(components.ExtSym)
		err = binary.Read(metRdr, binary.LittleEndian, &s.SymBase)
		if err != nil {
			return nil, err
		}
		s.Path = make([]byte, s.PathSize)
		err = binary.Read(metRdr, binary.LittleEndian, &s.Path)
		if err != nil {
			return nil, err
		}
		err = binary.Read(metRdr, binary.LittleEndian, &s.XattrIndex)
		if err != nil {
			return nil, err
		}
	default:
		err = binary.Read(metRdr, binary.LittleEndian, &inode.Data)
		if err != nil {
			return nil, err
		}
	}
	return inode, nil
}
