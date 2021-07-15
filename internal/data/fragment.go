package data

import (
	"fmt"
	"io"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/decompress"
)

type Fragment struct {
	entry  components.FragBlockEntry
	offset uint32
	size   uint32
}

func (f Fragment) GetDataReader(rdr io.ReaderAt, decomp decompress.Decompressor) (io.ReadCloser, error) {
	fmt.Println("HI", f.offset)
	return GetDataBlockReader(rdr, f.entry.Start, f.offset, f.entry.Size, decomp, f.size)
}
