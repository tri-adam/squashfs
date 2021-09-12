package squashfs

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"

	"github.com/CalebQ42/squashfs/internal/components"
)

//FS is a directory in the squashfs. Implements fs.FS.
type FS struct {
	r       *Reader
	parent  *FS
	entries []dirEntry
}

func (r *Reader) fsFromInode(i *components.Inode, parent *FS) (*FS, error) {
	ents, err := r.getDirEntriesFromInode(i)
	return &FS{
		entries: ents,
		r:       r,
		parent:  parent,
	}, err
}

//Open returns a *squashfs.File cast as a fs.File.
//Accepts wildcards, but cannot look up the director tree (../)
func (f FS) Open(filepath string) (fs.File, error) {
	e, err := f.getEntAt(filepath)
	if err != nil {
		return nil, err
	}
	return f.r.fileFromEntry(e, &f)
}

//ExtractTo extracts the directory to the given filepath.
func (f FS) ExtractTo(filepath string) (err error) {
	return f.ExtractToOptions(filepath, DefaultOptions())
}

//ExtractToOptions extracts the directory to the given filepath with the given options.
func (f FS) ExtractToOptions(filepath string, options ExtractionOptions) (err error) {
	filepath = path.Clean(filepath)
	if options.Verbose {
		log.Println("Extracting folder to: ", filepath)
	}
	err = os.Mkdir(filepath, options.FolderPerm)
	if err != nil {
		if options.Verbose {
			log.Println("Error while creating folder: ", filepath)
		}
		return
	}
	errChan := make(chan error)
	for i := range f.entries {
		go func(e dirEntry) {
			subDir, er := f.r.fileFromEntry(e, &f)
			if er != nil {
				if options.Verbose {
					log.Println("Error while extracting file: ", e.Name(), er)
				}
				errChan <- er
				return
			}
			er = subDir.ExtractToOptions(filepath, options)
			if er != nil && options.Verbose {
				log.Println("Error while extracting file: ", e.Name(), er)
			}
			errChan <- er
		}(f.entries[i])
	}
	for i := 0; i < len(f.entries); i++ {
		err = <-errChan
		if err != nil {
			if !options.AllowErrors {
				return
			}
		}
	}
	return
}

//ReadDir returns the fs.DirEntry's contained is the given directory.
func (f FS) ReadDir(name string) (out []fs.DirEntry, err error) {
	var ents []dirEntry
	if name != "" && name != "./" {
		var e dirEntry
		e, err = f.getEntAt(name)
		if err != nil {
			return
		}
		var in *components.Inode
		in, err = f.r.dirEntryToInode(e)
		if err != nil {
			return
		}
		ents, err = f.r.getDirEntriesFromInode(in)
		if err != nil {
			return
		}
	} else {
		ents = f.entries
	}
	out = make([]fs.DirEntry, len(ents))
	for i, e := range ents {
		out[i] = e
	}
	return
}

//Stat returns the fs.FileInfo for the given file.
func (f FS) Stat(name string) (fs.FileInfo, error) {
	e, err := f.getEntAt(name)
	if err != nil {
		return nil, err
	}
	in, err := f.r.dirEntryToInode(e)
	if err != nil {
		return nil, err
	}
	return FileInfo{
		i:   in,
		r:   f.r,
		ent: e,
	}, nil
}

//Sub returns the fs.FS for the given directory.
func (f FS) Sub(dir string) (fs.FS, error) {
	e, err := f.getEntAt(dir)
	if err != nil {
		return nil, err
	}
	in, err := f.r.dirEntryToInode(e)
	if err != nil {
		return nil, err
	}
	return f.r.fsFromInode(in, &f)
}

func (f FS) getEntAt(filepath string) (dirEntry, error) {
	parts := strings.Split(filepath, "/")
	if parts[0] == ".." {
		if f.parent == nil {
			return dirEntry{}, errors.New("openning ../ on root")
		}
		return f.parent.getEntAt(strings.Join(parts[1:], "/"))
	}
	for _, e := range f.entries {
		if is, _ := path.Match(parts[0], string(e.ent.Name)); is {
			if len(parts) == 1 {
				return e, nil
			} else {
				in, err := f.r.dirEntryToInode(e)
				if err != nil {
					return dirEntry{}, err
				}
				fs, err := f.r.fsFromInode(in, &f)
				if err != nil {
					return dirEntry{}, err
				}
				return fs.getEntAt(strings.Join(parts[1:], "/"))
			}
		}
	}
	return dirEntry{}, fs.ErrNotExist
}
