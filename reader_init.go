package squashfs

import (
	"bytes"
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
	if r.super.ParseFlags().CompressorOptions {
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
			if left > 512 {
				read = left
			}
			_, err = r.rdr.Seek(int64(b), io.SeekStart)
			if err != nil {
				return err
			}
			tmp := make([]components.FragBlockEntry, read)
			var byts []byte
			byts, err = metadata.ReadMetadata(r.rdr, int64(b), 0, r.decomp, int64(binary.Size(tmp)))
			if err != nil {
				return err
			}
			err = binary.Read(bytes.NewReader(byts), binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
			r.fragEntries = append(r.fragEntries, tmp...)
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
			if left > 2048 {
				read = left
			}
			_, err = r.rdr.Seek(int64(b), io.SeekStart)
			if err != nil {
				return err
			}
			tmp := make([]uint32, read)
			var byts []byte
			byts, err = metadata.ReadMetadata(r.rdr, int64(b), 0, r.decomp, int64(binary.Size(tmp)))
			if err != nil {
				return err
			}
			err = binary.Read(bytes.NewReader(byts), binary.LittleEndian, &tmp)
			if err != nil {
				return err
			}
			r.ids = append(r.ids, tmp...)
		}
	}
	root, err := r.parseInodeRef(r.super.RootInodeRef)
	if err != nil {
		return err
	}
	_ = root
	//TODO: parse root as FS
	return nil
}
