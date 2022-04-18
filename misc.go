package squashfs

func (r Reader) parseReference(ref uint64) (offset uint64, meta uint64) {
	return (ref >> 16) + r.s.inodeTableStart, ref & 0xFFFF
}
