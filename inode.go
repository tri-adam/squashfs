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
	switch inode.Type {
	case components.ExtDirType:
		d := inode.Data.(components.ExtDir)
		for len(data) < 24 {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &d.ExtDirBase)
		if err != nil {
			return nil, err
		}
		data = data[24:]
		d.Indexes = make([]components.DirIndex, d.IndexCount)
		for i := range d.Indexes {
			for len(data) < 12 {
				block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
				if err != nil {
					return nil, err
				}
				data = append(data, block...)
			}
			err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &d.Indexes[i].DirIndexBase)
			if err != nil {
				return nil, err
			}
			data = data[12:]
			d.Indexes[i].Name = make([]byte, d.Indexes[i].NameSize)
			for len(data) < len(d.Indexes[i].Name) {
				block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
				if err != nil {
					return nil, err
				}
				data = append(data, block...)
			}
			err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &d.Indexes[i].Name)
			if err != nil {
				return nil, err
			}
			data = data[len(d.Indexes[i].Name):]
		}
	case components.FileType:
		f := inode.Data.(components.File)
		for len(data) < 16 {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &f.FileBase)
		if err != nil {
			return nil, err
		}
		data = data[16:]
		sizeNum := f.Size / r.super.BlockSize
		if f.Size&r.super.BlockSize > 0 {
			sizeNum++
		}
		f.BlockSizes = make([]uint32, sizeNum)
		for len(data) < 4*int(sizeNum) {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &f.BlockSizes)
		if err != nil {
			return nil, err
		}
	case components.ExtFileType:
		f := inode.Data.(components.ExtFile)
		for len(data) < 40 {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &f.ExtFileBase)
		if err != nil {
			return nil, err
		}
		data = data[40:]
		sizeNum := f.Size / uint64(r.super.BlockSize)
		if f.Size&uint64(r.super.BlockSize) > 0 {
			sizeNum++
		}
		f.BlockSizes = make([]uint32, sizeNum)
		for len(data) < 4*int(sizeNum) {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &f.BlockSizes)
		if err != nil {
			return nil, err
		}
	case components.SymType:
		s := inode.Data.(components.Sym)
		for len(data) < 8 {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &s.SymBase)
		if err != nil {
			return nil, err
		}
		data = data[8:]
		s.Path = make([]byte, s.PathSize)
		for len(data) < int(s.PathSize) {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &s.Path)
		if err != nil {
			return nil, err
		}
	case components.ExtSymType:
		s := inode.Data.(components.ExtSym)
		for len(data) < 8 {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &s.SymBase)
		if err != nil {
			return nil, err
		}
		data = data[8:]
		s.Path = make([]byte, s.PathSize)
		for len(data) < int(s.PathSize) {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &s.Path)
		if err != nil {
			return nil, err
		}
		data = data[s.PathSize:]
		for len(data) < 4 {
			block, nextBlock, err = metadata.ReadBlockAt(r.rdr, nextBlock, r.decomp)
			if err != nil {
				return nil, err
			}
			data = append(data, block...)
		}
		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &s.XattrIndex)
		if err != nil {
			return nil, err
		}
	default:
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
