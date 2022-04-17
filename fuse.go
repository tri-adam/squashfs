package squashfs

import (
	"context"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

type Fuse struct {
	fuseutil.NotImplementedFileSystem
	r       *Reader
	mount   *fuse.MountedFileSystem
	handles map[fuseops.HandleID]handler
}

func NewSquashfuse(r *Reader) *Fuse {
	return &Fuse{r: r}
}

func (f *Fuse) Mount(dir string) (err error) {
	options := &fuse.MountConfig{
		ReadOnly: true,
		Subtype:  "squashfs",
	}
	f.mount, err = fuse.Mount(dir, fuseutil.NewFileSystemServer(f), options)
	return
}

func (f *Fuse) UnMount() (err error) {
	return f.mount.Join(context.Background())
}
