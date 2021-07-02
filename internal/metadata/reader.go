package metadata

import (
	"encoding/binary"
	"io"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

//ReadMetadata reads len bytes from metadata blocks from rdr. Decompresses and handles metadata headers as needed.
func ReadMetadata(rdr io.Reader, decomp decompress.Decompressor, length int) (out []byte, err error) {
	var hdr uint16
	var tmpData []byte
	for length > 0 {
		err = binary.Read(rdr, binary.LittleEndian, &hdr)
		if err != nil {
			return
		}
		tmpData = make([]byte, hdr&0x7FFF)
		_, err = rdr.Read(tmpData)
		if err != nil {
			return
		}
		if hdr&0x8000 != 0x8000 {
			tmpData, err = decompress.Decompress(decomp, tmpData)
			if err != nil {
				return
			}
			out = append(out, tmpData...) //TODO: remove appends by presetting out length
		} else {
			out = append(out, tmpData...)
		}
		length -= len(tmpData)
		//CONTINUE HERE
	}
	return
}
