package decompress

import (
	"bytes"
	"compress/gzip"
	"io"
)

func Decompress(d Decompressor, data []byte) ([]byte, error) {
	dataReader := bytes.NewReader(data)
	rdr, err := d.Reader(dataReader)
	if err != nil {
		return nil, err
	}
	defer rdr.Close()
	return io.ReadAll(rdr)
}

type Decompressor interface {
	Reader(io.Reader) (io.ReadCloser, error)
}

type Gzip struct {
	CompressionLevel int32
	WindowSize       int16
	Strategies       int16
}

func (g *Gzip) Reader(rdr io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(rdr)
}
