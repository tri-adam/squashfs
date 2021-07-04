package metadata

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/CalebQ42/squashfs/internal/decompress"
)

//ReadMetadata reads metadata blocks.
//Offset is the offset to the start of the metadata block and blockOffset is the offset into the first block.
//Automatically reads through multiple blocks and decompresses as necessary.
//Returned out will ALWAYS be size in len, but not not be filled if err != nil.
func ReadMetadata(rdr io.ReaderAt, offset, blockOffset int64, decomp decompress.Decompressor, size int64) (out []byte, err error) {
	blocks := int(size+blockOffset) / 8192
	if (size+blockOffset)%8192 > 0 {
		blocks++
	}
	var sec *io.SectionReader
	var group sync.WaitGroup
	var num int
	var hdr uint16
	outChan := make(chan blockHandler)
	for num < blocks {
		sec = io.NewSectionReader(rdr, int64(offset), 2)
		err = binary.Read(sec, binary.LittleEndian, &hdr)
		group.Add(1)
		go readBlock(rdr, offset+2, hdr&0x7FFF, hdr&0x8000 != 0x8000, decomp, num, group, outChan)
		offset += int64(hdr&0x7FFF) + 2
		num++
		blocks--
	}
	num = 0
	var backlog []blockHandler
	for num < blocks && len(backlog) > 0 {
		handler := <-outChan
		if handler.err != nil {
			group.Wait()
			err = handler.err
			return
		}
		if handler.num != num {
			backlog = append(backlog, handler)
			continue
		}
		if num == 0 {
			out = append(out, handler.data[blockOffset:]...)
		} else {
			out = append(out, handler.data...)
		}
		num++
		for i := 0; i < len(backlog); i++ {
			if backlog[i].num == num {
				if num == 0 {
					out = append(out, handler.data[blockOffset:]...)
				} else {
					out = append(out, handler.data...)
				}
				num++
				i = -1
			}
		}
	}
	return
}

type blockHandler struct {
	err  error
	data []byte
	num  int
}

func readBlock(rdr io.ReaderAt, offset int64, size uint16, compressed bool, decomp decompress.Decompressor, num int, group sync.WaitGroup, out chan blockHandler) {
	defer group.Done()
	handler := blockHandler{
		num: num,
	}
	handler.data = make([]byte, size)
	_, handler.err = rdr.ReadAt(handler.data, offset)
	if handler.err != nil {
		out <- handler
		group.Done()
	}
	if compressed {
		handler.data, handler.err = decompress.Decompress(decomp, handler.data)
	}
	out <- handler
}

//ReadAt reads a single metadta block at the offset.
func ReadBlockAt(rdr io.ReaderAt, offset int64, decomp decompress.Decompressor) (out []byte, nextBlock int64, err error) {
	hdrRdr := io.NewSectionReader(rdr, offset, 2)
	var hdr uint16
	err = binary.Read(hdrRdr, binary.LittleEndian, &hdr)
	if err != nil {
		return
	}
	nextBlock = offset + 2 + int64(hdr&0x7FFF)
	out = make([]byte, hdr&0x7FFF)
	_, err = rdr.ReadAt(out, offset+2)
	if err != nil {
		return
	}
	if hdr&0x8000 != 0x8000 {
		if decomp == nil {
			return nil, 0, errors.New("compressed metadata block, but no decompressor given")
		}
		out, err = decompress.Decompress(decomp, out)
	}
	return
}
