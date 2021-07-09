package squashfs

import (
	"io/fs"

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
		f.rdr, err = data.NewReaderFromInode(f.r.rdr, f.r.decomp, f.i, f.r.fragTable)
		if err != nil {
			return
		}
	}
	return f.rdr.Read(p)
}

func (f File) Close() error {
	if f.rdr != nil {
		return f.rdr.Close()
	}
	return nil
}

func (f File) ExtractTo(filepath string) error {
	//TODO
	return nil
}
