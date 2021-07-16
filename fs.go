package squashfs

import (
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/CalebQ42/squashfs/internal/components"
)

type FS struct {
	r       *Reader
	entries []dirEntry
}

func (r Reader) fsFromInode(i *components.Inode) (FS, error) {
	ents, err := r.getDirEntriesFromInode(i)
	return FS{
		entries: ents,
		r:       &r,
	}, err
}

func (f FS) Open(filepath string) (fs.File, error) {
	parts := strings.Split(filepath, "/")
	for _, e := range f.entries {
		if is, _ := path.Match(parts[0], string(e.Name)); is {
			if len(parts) == 1 {
				return f.r.FileFromEntry(e)
			} else {
				in, err := f.r.dirEntryToInode(&e)
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

func (f FS) ExtractTo(filepath string) (err error) {
	filepath = path.Clean(filepath)
	err = os.Mkdir(filepath, os.ModePerm)
	if err != nil {
		return
	}
	errChan := make(chan error)
	for _, e := range f.entries {
		go func(e dirEntry) {
			subDir, er := f.r.FileFromEntry(e)
			if er != nil {
				errChan <- er
				return
			}
			errChan <- subDir.ExtractTo(filepath)
			subDir.Close()
		}(e)
	}
	for i := 0; i < len(f.entries); i++ {
		err = <-errChan
		if err != nil {
			return
		}
	}
	return
}
