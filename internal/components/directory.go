package components

type DirHeader struct {
	Count    uint32
	Start    uint32
	InodeNum uint32
}

type DirEntryBase struct {
	Offset      uint16
	InodeOffset int16
	Type        uint16
	NameSize    uint16
}

type DirEntry struct { //ignore fieldalignment warning
	DirEntryBase
	Name []byte
}
