package squashfs

import "io/fs"

type DirEntry struct{}

//Info returns the fs.FileInfo for the file pointed to by the DirEntry
func (d DirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
	//TODO
}

//IsDir does exactly what you think it does
func (d DirEntry) IsDir() bool {
	return false
	//TODO
}

func (d DirEntry) Name() string {
	return ""
	//TODO
}

func (d DirEntry) Type() fs.FileMode {
	return 0
	//TODO
}
