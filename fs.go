package squashfs

import "github.com/CalebQ42/squashfs/internal/components"

type FS struct {
	entries []components.DirEntry
}

// func (f FS) Open(path string) (fs.File, error){}
