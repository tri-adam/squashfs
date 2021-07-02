package components

const SuperblockSize = 96 //bytes
const SuperblockMagic = 0x73717368

type Superblock struct {
	Magic             uint32
	InodeCount        uint32
	ModTime           uint32
	BlockSize         uint32
	FragEntryCount    uint32
	CompressionID     uint16
	BlockLog          uint16
	Flags             uint16
	IDCount           uint16
	VersionMajor      uint16
	VersionMinor      uint16
	RootInodeRef      uint64
	BytesUsed         uint64
	IDTableStart      uint64
	XattrIDTableStart uint64
	InodeTableStart   uint64
	DirTableStart     uint64
	FragTableStart    uint64
	ExportTableStart  uint64
}

type SuperblockFlags struct {
	UncompressedInodes bool
	UncompressedData   bool
	Check              bool
	UncompressedFrags  bool
	NoFrags            bool
	AlwaysFrag         bool
	Duplicates         bool
	Exportable         bool
	UncompressedXattrs bool
	NoXattrs           bool
	CompressorOptions  bool
	UncompressedIDs    bool
}

func (s Superblock) ParseFlags() SuperblockFlags {
	return SuperblockFlags{
		UncompressedInodes: s.Flags&0x0001 == 0x0001,
		UncompressedData:   s.Flags&0x0002 == 0x0002,
		Check:              s.Flags&0x0004 == 0x0004,
		UncompressedFrags:  s.Flags&0x0008 == 0x0008,
		NoFrags:            s.Flags&0x0010 == 0x0010,
		AlwaysFrag:         s.Flags&0x0020 == 0x0020,
		Duplicates:         s.Flags&0x0040 == 0x0040,
		Exportable:         s.Flags&0x0080 == 0x0080,
		UncompressedXattrs: s.Flags&0x0100 == 0x0100,
		NoXattrs:           s.Flags&0x0200 == 0x0200,
		CompressorOptions:  s.Flags&0x0400 == 0x0400,
		UncompressedIDs:    s.Flags&0x0400 == 0x0400,
	}
}
