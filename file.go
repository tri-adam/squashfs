package squashfs

import (
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/data"
)

//File is a file within a squashfs archive. Implements fs.File
type File struct {
	i   *components.Inode
	rdr *data.Reader
	r   *Reader
	ent dirEntry

	parent *FS
}

func (r Reader) fileFromEntry(d dirEntry, parent *FS) (*File, error) {
	i, err := r.dirEntryToInode(d)
	if err != nil {
		return nil, err
	}
	return &File{
		i:      i,
		r:      &r,
		ent:    d,
		parent: parent,
	}, nil
}

//Stat returns a file's fs.FileInfo
func (f File) Stat() (fs.FileInfo, error) {
	return FileInfo{
		i:   f.i,
		ent: f.ent,
		r:   f.r,
	}, nil
}

//Read reads p bytes from the file. If called after Close, the reader is re-created.
func (f *File) Read(p []byte) (n int, err error) {
	if f.rdr == nil {
		f.rdr, err = data.NewReaderFromInode(f.r.rdr, f.r.super.BlockSize, f.r.decomp, f.i, f.r.fragTable)
		if err != nil {
			return
		}
	}
	return f.rdr.Read(p)
}

//WriteTo writes all data from the File to the io.Writer.
//Creates a new reader so Read calls are uneffected.
func (f File) WriteTo(w io.Writer) (n int64, err error) {
	rdr, err := data.NewReaderFromInode(f.r.rdr, f.r.super.BlockSize, f.r.decomp, f.i, f.r.fragTable)
	if err != nil {
		return
	}
	defer rdr.Close()
	return rdr.WriteTo(w)
}

//Close closes the file's reader. Subsequent Read calls will re-create the reader.
func (f *File) Close() error {
	if f.rdr != nil {
		return f.rdr.Close()
	}
	f.rdr = nil
	return nil
}

//FS returns the given File as a squashfs.FS
func (f File) FS() (*FS, error) {
	return f.r.fsFromInode(f.i)
}

//GetSymlinkFile returns the linked File if the given file is a symlink.
//If not, this simply returns the calling File.
func (f File) GetSymlinkFile() *File {
	return nil
	//TODO
}

//IsDir is exactly what you think it is.
func (f File) IsDir() bool {
	return false
	//TODO
}

//IsSymlink returns if the file is a symlink, and if so returns the path it's pointed to.
func (f File) IsSymlink() (bool, string) {
	return false, ""
	//TODO
}

//ReadDir returns n fs.DirEntries contianed in the File (if it is a directory).
//If n <= 0 all fs.DirEntry's are returned.
func (f File) ReadDir(n int) ([]fs.DirEntry, error) {
	return nil, nil
	//TODO.
}

//ExtractTo extract the given File to the given location.
func (f *File) ExtractTo(filepath string) (err error) {
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
		_, err = io.Copy(fil, f)
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
		errChan := make(chan error)
		for i := range entries {
			go func(e dirEntry) {
				subDir, er := f.r.fileFromEntry(e, nil)
				if er != nil {
					errChan <- er
					return
				}
				er = subDir.ExtractTo(filepath)
				errChan <- er
			}(entries[i])
		}
		for i := 0; i < len(entries); i++ {
			err = <-errChan
			if err != nil {
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
