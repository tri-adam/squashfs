package directory

import (
	"encoding/binary"
	"io"
)

type Header struct {
	Entries    uint32
	InodeStart uint32
	Num        uint32
}

type entryInit struct {
	Offset    uint16
	NumOffset int16
	Type      uint16
	nameSize  uint16
}

type Entry struct {
	entryInit
	Name   []byte
	Header *Header
}

type AllEntries []Entry

func readEntry(r io.Reader) (e Entry, err error) {
	err = binary.Read(r, binary.LittleEndian, &e.entryInit)
	if err != nil {
		return
	}
	e.Name = make([]byte, e.nameSize+1)
	err = binary.Read(r, binary.LittleEndian, &e.Name)
	return
}

func ReadEntries(r io.Reader, size uint32) (e AllEntries, err error) {

}
