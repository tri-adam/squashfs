package squashfs

import (
	"io"
	"io/fs"
	"os"
	"path"
	"sync"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/data"
)

type File struct {
	i   *components.Inode
	rdr *data.Reader
	r   *Reader
	ent dirEntry
}

func (r Reader) FileFromEntry(d dirEntry) (out File, err error) {
	out.ent = d
	out.i, err = r.dirEntryToInode(&d)
	out.r = &r
	return
}

func (f File) Stat() (fs.FileInfo, error) {
	return FileInfo{
		i:   f.i,
		ent: f.ent,
		r:   f.r,
	}, nil
}

func (f File) Read(p []byte) (n int, err error) {
	if f.rdr == nil {
		f.rdr, err = data.NewReaderFromInode(f.r.rdr, f.r.super.BlockSize, f.r.decomp, f.i, f.r.fragTable)
		if err != nil {
			return
		}
	}
	return f.rdr.Read(p)
}

func (f File) WriteTo(w io.Writer) (n int64, err error) {
	if f.rdr == nil {
		f.rdr, err = data.NewReaderFromInode(f.r.rdr, f.r.super.BlockSize, f.r.decomp, f.i, f.r.fragTable)
		if err != nil {
			return
		}
	}
	return f.rdr.WriteTo(w)
}

func (f File) Close() error {
	if f.rdr != nil {
		return f.rdr.Close()
	}
	return nil
}

func (f File) ExtractTo(filepath string) (err error) {
	filepath = path.Clean(filepath)
	os.Mkdir(filepath, os.ModePerm)
	filepath += "/" + string(f.ent.Name)
	var fil *os.File
	switch f.ent.Type {
	case components.FileType:
		os.Remove(filepath)
		fil, err = os.Create(filepath)
		if err != nil {
			return
		}
		var rdr *data.Reader
		rdr, err = data.NewReaderFromInode(f.r.rdr, f.r.super.BlockSize, f.r.decomp, f.i, f.r.fragTable)
		if err != nil {
			return
		}
		_, err = io.Copy(fil, rdr)
		if err != nil {
			return
		}
		fil.Chown(int(f.r.idTable[f.i.UIDIndex]), int(f.r.idTable[f.i.GIDIndex])) //don't report errors because those can happen often
	case components.SymType:
		os.Remove(filepath)
		if f.i.Type == components.SymType {
			err = os.Symlink(string(f.i.Data.(components.Sym).Path), filepath)
			if err != nil {
				return
			}
		} else if f.i.Type == components.ExtSymType {
			err = os.Symlink(string(f.i.Data.(components.ExtSym).Path), filepath)
			if err != nil {
				return
			}
		}
	case components.DirType:
		err = os.Mkdir(filepath, os.ModePerm)
		if err != os.ErrExist && err != nil {
			return
		}
		var entries []dirEntry
		entries, err = f.r.getDirEntriesFromInode(f.i)
		if err != nil {
			return
		}
		var group sync.WaitGroup
		errChan := make(chan error)
		for _, e := range entries {
			group.Add(1)
			go func(e dirEntry) {
				defer group.Done()
				subDir, er := f.r.FileFromEntry(e)
				if er != nil {
					errChan <- er
					return
				}
				er = subDir.ExtractTo(filepath)
				errChan <- er
			}(e)
		}
		for i := 0; i < len(entries); i++ {
			err = <-errChan
			if err != nil {
				// group.Wait()
				return
			}
		}
		fil, err = os.Open(filepath)
		if err != nil {
			return
		}
		fil.Chown(int(f.r.idTable[f.i.UIDIndex]), int(f.r.idTable[f.i.GIDIndex])) //don't report errors because those can happen often
	default:
		return //can only extract dir, sym, and regular. If not of those types, just gracefully ignore.
	}
	return
}
