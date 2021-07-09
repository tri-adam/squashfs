package squashfs

import (
	"io/fs"
	"time"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/data"
)

type FileInfo struct {
	i   *components.Inode
	r   *Reader
	ent dirEntry
}

func (f FileInfo) Name() string {
	return string(f.ent.Name)
}

func (f FileInfo) Size() int64 {
	if f.i.Type == components.FileType {
		return int64(f.i.Data.(components.File).Size)
	} else if f.i.Type == components.ExtFileType {
		return int64(f.i.Data.(components.ExtFile).Size)
	}
	return 0
}

func (f FileInfo) Mode() (out fs.FileMode) {
	out = fs.FileMode(f.i.Permissions)
	switch f.ent.Type {
	case components.DirType:
		out |= fs.ModeDir
	case components.SymType:
		out |= fs.ModeSymlink
	case components.CharType:
		out |= fs.ModeCharDevice
	case components.SocketType:
		out |= fs.ModeSocket
	}
	return
}

func (f FileInfo) ModTime() time.Time {
	return time.Unix(int64(f.i.ModTime), 0)
}

func (f FileInfo) IsDir() bool {
	return f.ent.Type == components.DirType
}

//Sys will try to return an io.ReadCloser for regular files and nil for other file types.
func (f FileInfo) Sys() interface{} {
	if f.ent.Type == components.FileType {
		rdr, err := data.NewReaderFromInode(f.r.rdr, f.r.decomp, f.i, f.r.fragTable)
		if err != nil {
			return nil
		}
		return rdr
	}
	return nil
}
