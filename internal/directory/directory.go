package directory

import (
	"encoding/binary"
	"io"
)

type header struct {
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

type entry struct {
	entryInit
	Name []byte
}

type Entry struct {
	Name        string
	BlockOffset uint32
	Type        uint16
	Offset      uint16
}

func readEntry(r io.Reader) (e entry, err error) {
	err = binary.Read(r, binary.LittleEndian, &e.entryInit)
	if err != nil {
		return
	}
	e.Name = make([]byte, e.nameSize+1)
	err = binary.Read(r, binary.LittleEndian, &e.Name)
	return
}

func ReadEntries(r io.Reader, size uint32) (e []Entry, err error) {
	e = make([]Entry, 0)
	readTotal := uint32(3)
	var h header
	var en entry
	for readTotal < size {
		err = binary.Read(r, binary.LittleEndian, &h)
		if err == io.EOF {
			continue
		} else if err != nil {
			return
		}
		readTotal += 12
		for i := uint32(0); i < h.Entries; i++ {
			en, err = readEntry(r)
			if err != nil {
				return
			}
			readTotal += 8 + uint32(en.nameSize)
			e = append(e, Entry{
				Name:        string(en.Name),
				Type:        en.Type,
				BlockOffset: h.InodeStart,
				Offset:      en.Offset,
			})
		}
	}
	return
}
