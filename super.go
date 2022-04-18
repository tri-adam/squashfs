package squashfs

import "math"

type superblock struct {
	magic            uint32
	inodeCount       uint32
	modTime          uint32
	blockSize        uint32
	fragCount        uint32
	compType         uint16
	blockLog         uint16
	flags            uint16
	idCount          uint16
	verMaj           uint16
	verMin           uint16
	rootInodeRef     uint64
	size             uint64
	idTableStart     uint64
	xattrTableStart  uint64
	inodeTableStart  uint64
	dirTableStart    uint64
	fragTableStart   uint64
	exportTableStart uint64
}

func (s superblock) hasMagic() bool {
	return s.magic == 0x73717368
}

func (s superblock) checkBlockLog() bool {
	return s.blockLog == uint16(math.Log2(float64(s.blockSize)))
}

func (s superblock) uncompressedInodes() bool {
	return s.flags&0x1 == 0x1
}

func (s superblock) uncompressedData() bool {
	return s.flags&0x2 == 0x2
}
func (s superblock) uncompressedFragments() bool {
	return s.flags&0x8 == 0x8
}

func (s superblock) noFragments() bool {
	return s.flags&0x10 == 0x10
}

func (s superblock) alwaysFragment() bool {
	return s.flags&0x20 == 0x20
}

func (s superblock) duplicates() bool {
	return s.flags&0x40 == 0x40
}

func (s superblock) exportable() bool {
	return s.flags&0x80 == 0x80
}

func (s superblock) uncompressedXattrs() bool {
	return s.flags&0x100 == 0x100
}

func (s superblock) noXattrs() bool {
	return s.flags&0x200 == 0x200
}

func (s superblock) compressionOptions() bool {
	return s.flags&0x400 == 0x400
}

func (s superblock) uncompressedIDs() bool {
	return s.flags&0x800 == 0x800
}
