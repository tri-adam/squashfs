package decompress

import (
	"bytes"
	"compress/zlib"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/rasky/go-lzo"
	"github.com/therootcompany/xz"
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

func (g Gzip) Reader(rdr io.Reader) (io.ReadCloser, error) {
	return zlib.NewReader(rdr)
}

type Xz struct {
	DictionarySize int32
	ExecFilters    int32
}

func (x Xz) Reader(rdr io.Reader) (io.ReadCloser, error) {
	xz, err := xz.NewReader(rdr, uint32(x.DictionarySize))
	return io.NopCloser(xz), err
}

type Lz4 struct {
	Version int32
	Flags   int32
}

func (l Lz4) Reader(rdr io.Reader) (io.ReadCloser, error) {
	lz := lz4.NewReader(rdr)
	return io.NopCloser(lz), nil
}

type Zstd struct {
	CompressionLevel int32
}

func (z Zstd) Reader(rdr io.Reader) (io.ReadCloser, error) {
	zs, err := zstd.NewReader(rdr)
	return zs.IOReadCloser(), err
}

type Lzo struct {
	Algorithm int32
	Level     int32
}

func (l Lzo) Reader(rdr io.Reader) (io.ReadCloser, error) {
	byt, err := lzo.Decompress1X(rdr, 0, 0)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(byt)), nil
}
