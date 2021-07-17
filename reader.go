package squashfs

import (
	"io"
	"time"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/decompress"
	"github.com/CalebQ42/squashfs/internal/rawreader"
)

type Reader struct {
	*FS

	rdr rawreader.RawReader

	decomp    decompress.Decompressor
	fragTable []components.FragBlockEntry
	idTable   []uint32
	super     components.Superblock
}

//NewReader creates a new squashfs.Reader from an io.ReaderAt.
func NewReader(reader io.ReaderAt) (*Reader, error) {
	out := Reader{
		rdr: rawreader.ConvertReaderAt(reader),
	}
	err := out.init()
	if err != nil {
		return nil, err
	}
	return &out, nil
}

//NewReaderFromReader creates a new squashfs.Reader from an io.Reader.
//If the reader implements io.Seeker then that is used, otherwise data is cached to prevent it from being lost.
//This is not ideal (as I haven't implemented removing data once it's been used) and io.ReadSeeker or io.ReaderAt is prefered.
func NewReaderFromReader(reader io.Reader) (*Reader, error) {
	out := Reader{
		rdr: rawreader.ConvertReader(reader),
	}
	err := out.init()
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *Reader) ModTime() time.Time {
	return time.Unix(int64(r.super.ModTime), 0)
}

// func (r Reader) Export(path string) error {}

// func (r Reader) ListAllFiles() ([]string, error) {}
