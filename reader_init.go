package squashfs

import (
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/CalebQ42/squashfs/internal/components"
	"github.com/CalebQ42/squashfs/internal/decompress"
	"github.com/CalebQ42/squashfs/internal/metadata"
)

func (r *Reader) init() error {
	//TODO
	err := binary.Read(r.rdr, binary.LittleEndian, &r.super)
	if err != nil {
		return err
	}
	if r.super.Magic != components.SuperblockMagic {
		return errors.New("magic number doesn't match. Reader might not be a squashfs archive or corrupted")
	}
	if r.super.BlockLog != uint16(math.Log2(float64(r.super.BlockSize))) {
		return errors.New("blockLog and blockSize don't match. Reader might be corrupted")
	}
	if r.super.ParseFlags().CompressorOptions {
		switch r.super.CompressionID {
		case 1:
			r.decomp = decompress.Gzip{}
		case 3:
			r.decomp = decompress.Lzo{}
		case 4:
			r.decomp = decompress.Xz{}
		case 5:
			r.decomp = decompress.Lz4{}
		case 6:
			r.decomp = decompress.Zstd{}
		default:
			return errors.New("unsupported compression type")
		}
		err = binary.Read(r.rdr, binary.LittleEndian, r.decomp)
		if err != nil {
			return err
		}
		if r.super.CompressionID == 3 {
			lzo := r.decomp.(decompress.Lzo)
			if lzo.Algorithm != 0 && lzo.Algorithm != 4 {
				return errors.New("unsupported lzo compression algorithm")
			}
		}
	}
	if r.super.FragEntryCount > 0 {
		//fragment table
		count := r.super.FragEntryCount / 512
		if r.super.FragEntryCount%512 > 0 {
			count++
		}
		_, err = r.rdr.Seek(int64(r.super.FragTableStart), io.SeekStart)
		if err != nil {
			return err
		}
		entryOffsets := make([]uint64, count)
		err = binary.Read(r.rdr, binary.LittleEndian, &entryOffsets)
		if err != nil {
			return err
		}
		r.fragEntries = make([]components.FragBlockEntry, r.super.FragEntryCount)
		left := r.super.FragEntryCount
		var offsetNum int
		for left > 0 {
			read := uint32(512)
			if left > 512 {
				read = left
			}
			var out []byte
			out, err = metadata.ReadMetadata(r.rdr, int64(entryOffsets[offsetNum]), 0, r.decomp, int64(binary.Size(r.fragEntries[0]))*int64(read))
			if err != nil {
				return err
			}
			offsetNum++
		}
		//TODO: decide to save entry offsets or frag entriess
	}
	if r.super.IDCount > 0 {
		//ID table
	}
	//parse root inode
	return nil
}
