package squashfs

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/data"
)

//File is a file within a squashfs archive. Implements fs.File
type File struct {
	i      *components.Inode
	rdr    *data.Reader
	r      *Reader
	parent *FS
	ent    dirEntry
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
	return f.r.fsFromInode(f.i, f.parent)
}

//GetSymlinkFile returns the linked File if the given file is a symlink and the file is present in the archive.
//If not, this simply returns nil.
func (f File) GetSymlinkFile() *File {
	if !f.IsSymlink() {
		return nil
	}
	fil, err := f.parent.Open(f.SymlinkPath())
	if err != nil {
		return nil
	}
	return fil.(*File)
}

//IsDir is exactly what you think it is.
func (f File) IsDir() bool {
	return f.ent.ent.Type == components.DirType
}

//IsSymlink returns if the file is a symlink, and if so returns the path it's pointed to.
func (f File) IsSymlink() bool {
	return f.ent.ent.Type == components.SymType
}

//SymlinkPath returns the path the symlink is pointing to.
//Returns an empty string if not a symlink.
func (f File) SymlinkPath() string {
	if f.i.Type == components.SymType {
		return string(f.i.Data.(components.Sym).Path)
	} else if f.i.Type == components.ExtSymType {
		return string(f.i.Data.(components.ExtSym).Path)
	}
	return ""
}

//ReadDir returns n fs.DirEntries contianed in the File (if it is a directory).
//If n <= 0 all fs.DirEntry's are returned.
func (f File) ReadDir(n int) (out []fs.DirEntry, err error) {
	ents, err := f.r.getDirEntriesFromInode(f.i)
	if err != nil {
		return nil, err
	}
	out = make([]fs.DirEntry, len(ents))
	for i, e := range ents {
		out[i] = e
	}
	return
}

//ExtractTo extracts the given File to the given location.
func (f File) ExtractTo(filepath string) (err error) {
	return f.ExtractToOptions(filepath, DefaultOptions())
}

//ExtractToOptions extracts the given File to the given location with the given options.
func (f File) ExtractToOptions(filepath string, options ExtractionOptions) (err error) {
	filepath = path.Clean(filepath)
	os.Mkdir(filepath, options.FolderPerm)
	filepath += "/" + string(f.ent.ent.Name)
	if options.Verbose {
		log.Println("Extracting to: ", filepath)
	}
	var fil *os.File
	switch f.ent.ent.Type {
	case components.FileType:
		os.Remove(filepath)
		fil, err = os.Create(filepath)
		if err != nil {
			if options.Verbose {
				log.Println("Error while creating file: ", err)
			}
			if !options.AllowErrors {
				return
			}
		}
		var rdr *data.Reader
		rdr, err = data.NewReaderFromInode(f.r.rdr, f.r.super.BlockSize, f.r.decomp, f.i, f.r.fragTable)
		if err != nil {
			if options.Verbose {
				log.Println("Error while making data reader: ", err)
			}
			if !options.AllowErrors {
				return
			}
		}
		defer rdr.Close()
		_, err = io.Copy(fil, rdr)
		if err != nil {
			if options.Verbose {
				log.Println("Error while copying data: ", err)
			}
			if !options.AllowErrors {
				return
			}
		}
		fil.Chown(int(f.r.idTable[f.i.UIDIndex]), int(f.r.idTable[f.i.GIDIndex])) //don't report errors because those can happen often
	case components.SymType:
		os.Remove(filepath)
		if options.DereferenceSymlink {
			symFil := f.GetSymlinkFile()
			if symFil == nil {
				if options.Verbose {
					log.Println("Symlink's target not available to dereference")
				}
				goto makeSym
			}
			symFil.ent.ent.Name = f.ent.ent.Name
			err = symFil.ExtractToOptions(path.Dir(filepath), options)
			if err != nil {
				if options.Verbose {
					log.Println("Error while extracting dereferenced symlink: ", err)
				}
				if !options.AllowErrors {
					return
				}
			}
			fil, err = os.Open(filepath)
			if err != nil {
				if options.Verbose {
					log.Println("Error while opening dereferenced symlink to set owner and permission: ", err)
				}
				if !options.AllowErrors {
					return
				}
			}
			fil.Chmod(fs.FileMode(f.i.Permissions))
			fil.Chown(int(f.r.idTable[f.i.UIDIndex]), int(f.r.idTable[f.i.GIDIndex])) //don't report errors because those can happen often
			return
		}
	makeSym:
		err = os.Symlink(f.SymlinkPath(), filepath)
		if err != nil {
			if options.Verbose {
				log.Println("Error while symlinking: ", err)
			}
			if !options.AllowErrors {
				return
			}
		}
		fil, err = os.Open(filepath)
		if err != nil {
			if options.Verbose {
				log.Println("Error while opening symlink to set owner and permission: ", err)
			}
			if !options.AllowErrors {
				return
			}
		}
		fil.Chmod(fs.FileMode(f.i.Permissions))
		fil.Chown(int(f.r.idTable[f.i.UIDIndex]), int(f.r.idTable[f.i.GIDIndex])) //don't report errors because those can happen often
		if !options.DereferenceSymlink && options.UnbreakSymlink {
			symPath := path.Dir(filepath) + f.SymlinkPath()
			_, err = os.Open(symPath)
			if os.IsNotExist(err) {
				symFil := f.GetSymlinkFile()
				if symFil == nil {
					if options.Verbose {
						log.Println("Symlink's target not available to unbreak")
					}
					return
				}
				return symFil.ExtractToOptions(symPath, options)
			}
		}
	case components.DirType:
		var fsFil *FS
		fsFil, err = f.FS()
		if err != nil {
			if options.Verbose {
				log.Println("Error while getting directory info: ", err)
			}
			if !options.AllowErrors {
				return
			}
		}
		err = fsFil.ExtractToOptions(filepath, options)
		if err != nil {
			if options.Verbose {
				log.Println("Error while extracting directory files: ", err)
			}
			if !options.AllowErrors {
				return
			}
		}
		fil, err = os.Open(filepath)
		if err != nil {
			if options.Verbose {
				log.Println("Error while opening folder to set owner and permission: ", err)
			}
			if !options.AllowErrors {
				return
			}
		}
		fil.Chmod(fs.FileMode(f.i.Permissions))
		fil.Chown(int(f.r.idTable[f.i.UIDIndex]), int(f.r.idTable[f.i.GIDIndex])) //don't report errors because those can happen often
	default:
		return //can only extract dir, sym, and regular. If not of those types, just gracefully ignore.
	}
	return
}
