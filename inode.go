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
	metRdr, err := metadata.NewReader(r.rdr, offset+r.super.InodeTableStart, blockOffset, r.decomp)
	if err != nil {
		return nil, err
	}
	err = binary.Read(metRdr, binary.LittleEndian, &inode.InodeHeader)
	if err != nil {
		return nil, err
	}
	switch inode.Type {
	case components.DirType:
		d := components.Dir{}
		err = binary.Read(metRdr, binary.LittleEndian, &d)
		if err != nil {
			return nil, err
		}
		inode.Data = d
	case components.FileType:
		f := components.File{}
		err = binary.Read(metRdr, binary.LittleEndian, &f.FileBase)
		if err != nil {
			return nil, err
		}
		sizeNum := f.Size / r.super.BlockSize
		if f.FragIndex == 0xFFFFFFFF {
			if f.Size%r.super.BlockSize > 0 {
				sizeNum++
			}
		}
		f.BlockSizes = make([]uint32, sizeNum)
		err = binary.Read(metRdr, binary.LittleEndian, &f.BlockSizes)
		if err != nil {
			return nil, err
		}
		inode.Data = f
	case components.SymType:
		s := components.Sym{}
		err = binary.Read(metRdr, binary.LittleEndian, &s.SymBase)
		if err != nil {
			return nil, err
		}
		s.Path = make([]byte, s.PathSize)
		err = binary.Read(metRdr, binary.LittleEndian, &s.Path)
		if err != nil {
			return nil, err
		}
		inode.Data = s
	case components.BlockType:
		fallthrough
	case components.CharType:
		d := components.Device{}
		err = binary.Read(metRdr, binary.LittleEndian, &d)
		if err != nil {
			return nil, err
		}
		inode.Data = d
	case components.FifoType:
		fallthrough
	case components.SocketType:
		d := components.IPC{}
		err = binary.Read(metRdr, binary.LittleEndian, &d)
		if err != nil {
			return nil, err
		}
		inode.Data = d
	case components.ExtDirType:
		d := components.ExtDir{}
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
		inode.Data = d
	case components.ExtFileType:
		f := components.ExtFile{}
		err = binary.Read(metRdr, binary.LittleEndian, &f.ExtFileBase)
		if err != nil {
			return nil, err
		}
		sizeNum := f.Size / uint64(r.super.BlockSize)
		if f.FragIndex == 0xFFFFFFFF {
			if f.Size%uint64(r.super.BlockSize) > 0 {
				sizeNum++
			}
		}
		f.BlockSizes = make([]uint32, sizeNum)
		err = binary.Read(metRdr, binary.LittleEndian, &f.BlockSizes)
		if err != nil {
			return nil, err
		}
		inode.Data = f
	case components.ExtSymType:
		s := components.ExtSym{}
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
		inode.Data = s
	case components.ExtBlockType:
		fallthrough
	case components.ExtCharType:
		d := components.ExtDevice{}
		err = binary.Read(metRdr, binary.LittleEndian, &d)
		if err != nil {
			return nil, err
		}
		inode.Data = d
	case components.ExtFifoType:
		fallthrough
	case components.ExtSocketType:
		d := components.ExtIPC{}
		err = binary.Read(metRdr, binary.LittleEndian, &d)
		if err != nil {
			return nil, err
		}
		inode.Data = d
	default:
		return nil, errors.New("inode type is unsupported: " + strconv.Itoa(int(inode.Type)))
	}
	return inode, nil
}
