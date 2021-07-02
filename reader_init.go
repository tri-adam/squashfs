package squashfs

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/CalebQ42/squashfs/internal/components"
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
	return nil
}
