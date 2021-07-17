package squashfs

import (
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/CalebQ42/squashfs/internal/components"
)

//FS is a directory in the squashfs. Implements fs.FS
type FS struct {
	r       *Reader
	entries []dirEntry
}

func (r *Reader) fsFromInode(i *components.Inode) (*FS, error) {
	ents, err := r.getDirEntriesFromInode(i)
	return &FS{
		entries: ents,
		r:       r,
	}, err
}

//Open returns a *squashfs.File cast as a fs.File.
//Accepts wildcards, but cannot look up the director tree (../)
func (f FS) Open(filepath string) (fs.File, error) {
	parts := strings.Split(filepath, "/")
	for _, e := range f.entries {
		if is, _ := path.Match(parts[0], string(e.Name)); is {
			if len(parts) == 1 {
				return f.r.fileFromEntry(e)
			} else {
				in, err := f.r.dirEntryToInode(e)
				if err != nil {
					return nil, err
				}
				fs, err := f.r.fsFromInode(in)
				if err != nil {
					return nil, err
				}
				return fs.Open(strings.Join(parts[1:], "/"))
			}
		}
	}
	return nil, fs.ErrNotExist
}

//ExtractTo extracts the directory to the given filepath.
func (f FS) ExtractTo(filepath string) (err error) {
	filepath = path.Clean(filepath)
	err = os.Mkdir(filepath, os.ModePerm)
	if err != nil {
		return
	}
	errChan := make(chan error)
	for i := range f.entries {
		go func(e dirEntry) {
			subDir, er := f.r.fileFromEntry(e)
			if er != nil {
				errChan <- er
				return
			}
			errChan <- subDir.ExtractTo(filepath)
			subDir.Close()
		}(f.entries[i])
	}
	for i := 0; i < len(f.entries); i++ {
		err = <-errChan
		if err != nil {
			return
		}
	}
	return
}
