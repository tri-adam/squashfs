package compression

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/klauspost/compress/zlib"
)

//Gzip is a decompressor for gzip type compression. Uses zlib for compression and decompression
type Gzip struct {
	CompressionLevel int32
}

//NewGzipCompressorWithOptions creates a new gzip compressor/decompressor with options read from the given reader.
func NewGzipCompressorWithOptions(r io.Reader) (*Gzip, error) {
	var init struct {
		CompressionLevel int32
		WindowSize       int16
		Strategies       int16
	}
	err := binary.Read(r, binary.LittleEndian, &init)
	if err != nil {
		return nil, err
	}
	return &Gzip{
		CompressionLevel: init.CompressionLevel,
	}, nil
}

//Decompress reads the entirety of the given reader and returns it uncompressed as a byte slice.
func (g *Gzip) Decompress(r io.Reader) ([]byte, error) {
	rdr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	var data bytes.Buffer
	_, err = io.Copy(&data, rdr)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

//Compress compresses the given data (as a byte array) and returns the compressed data.
func (g *Gzip) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	var err error
	var w *zlib.Writer
	if g.CompressionLevel == 0 {
		w = zlib.NewWriter(&buf)
	} else {
		w, err = zlib.NewWriterLevel(&buf, int(g.CompressionLevel))
		if err != nil {
			return nil, err
		}
	}
	_, err = w.Write(data)
	w.Close()
	return buf.Bytes(), err
}
