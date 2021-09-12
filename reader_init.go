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
	switch r.super.CompressionID {
	case 1:
		var tmp decompress.Gzip
		if r.super.ParseFlags().CompressorOptions {
			err = binary.Read(r.rdr, binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
		}
		r.decomp = tmp
	case 2:
		r.decomp = decompress.Lzma{}
	case 3:
		var tmp decompress.Lzo
		if r.super.ParseFlags().CompressorOptions {
			err = binary.Read(r.rdr, binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
		}
		r.decomp = tmp
	case 4:
		var tmp decompress.Xz
		if r.super.ParseFlags().CompressorOptions {
			err = binary.Read(r.rdr, binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
		}
		r.decomp = tmp
	case 5:
		var tmp decompress.Lz4
		if r.super.ParseFlags().CompressorOptions {
			err = binary.Read(r.rdr, binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
		}
		r.decomp = tmp
	case 6:
		var tmp decompress.Zstd
		if r.super.ParseFlags().CompressorOptions {
			err = binary.Read(r.rdr, binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
		}
		r.decomp = tmp
	default:
		return errors.New("unsupported compression type")
	}
	var metRdr *metadata.Reader
	if r.super.FragEntryCount > 0 {
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
		left := r.super.FragEntryCount
		for _, b := range entryOffsets {
			read := uint32(512)
			if left < 512 {
				read = left
			}
			_, err = r.rdr.Seek(int64(b), io.SeekStart)
			if err != nil {
				return err
			}
			tmp := make([]components.FragBlockEntry, read)
			metRdr, err = metadata.NewReader(r.rdr, b, 0, r.decomp)
			if err != nil {
				return err
			}
			err = binary.Read(metRdr, binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
			r.fragTable = append(r.fragTable, tmp...)
			left -= read
		}
	}
	if r.super.IDCount > 0 {
		count := r.super.IDCount / 2048
		if r.super.IDCount%2048 > 0 {
			count++
		}
		_, err = r.rdr.Seek(int64(r.super.IDTableStart), io.SeekStart)
		if err != nil {
			return err
		}
		idOffsets := make([]uint64, count)
		err = binary.Read(r.rdr, binary.LittleEndian, &idOffsets)
		if err != nil {
			return err
		}
		left := r.super.IDCount
		for _, b := range idOffsets {
			read := uint16(2048)
			if left < 2048 {
				read = left
			}
			_, err = r.rdr.Seek(int64(b), io.SeekStart)
			if err != nil {
				return err
			}
			tmp := make([]uint32, read)
			err = metRdr.Reset(b, 0)
			if err != nil {
				return err
			}
			err = binary.Read(metRdr, binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
			r.idTable = append(r.idTable, tmp...)
			left -= read
		}
	}
	root, err := r.parseInodeRef(r.super.RootInodeRef)
	if err != nil {
		return err
	}
	r.FS, err = r.fsFromInode(root, nil)
	return err
}
