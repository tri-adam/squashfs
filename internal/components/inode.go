package components

const (
	//Inode types from InodeHeader.Type
	DirType = iota + 1
	FileType
	SymType
	BlockType
	CharType
	FifoType
	SocketType
	ExtDirType
	ExtFileType
	ExtSymType
	ExtBlockType
	ExtCharType
	ExtFifoType
	ExtSocketType
)

type Inode struct { //ignore fieldalignment warning.
	//Data holds one of the many types of Inodes.
	InodeHeader
	Data interface{}
}

type InodeHeader struct {
	Type        uint16
	Permissions uint16
	UIDIndex    uint16
	GIDIndex    uint16
	ModTime     uint32
	Num         uint32
}

type BasicDir struct {
	DirIndex       uint32
	HardLinks      uint32
	FileSize       uint16
	BlockOffset    uint16
	ParentInodeNum uint32
}

type ExtDirBase struct {
	HardLinks      uint32
	FileSize       uint32
	DirIndex       uint32
	ParentInodeNum uint32
	IndexCount     uint16
	BlockOffset    uint16
	XattrIndex     uint32
}

type ExtDir struct { //ignore fieldalignment warning
	ExtDirBase
	Indexes []DirIndex
}

type DirIndexBase struct {
	HeaderOffset uint32 //offset from the dir's first header to this header
	BlockStart   uint32
	NameSize     uint32
}

type DirIndex struct { //ignore fieldalignment warning
	DirIndexBase
	Name []byte
}

type FileBase struct {
	BlockStart      uint32
	FragIndex       uint32
	FragBlockOffset uint32
	Size            uint32
}

type File struct { //ignore fieldalignment warning
	FileBase
	BlockSizes []uint32
}

type ExtFileBase struct {
	BlockStart      uint64
	Size            uint64
	Sparse          uint64
	HardLinks       uint32
	FragIndex       uint32
	FragBlockOffset uint32
	XattrIndex      uint32
}

type ExtFile struct { //ignore fieldalignment warning
	ExtFileBase
	BlockSizes []uint32
}

type SymBase struct {
	HardLinks uint32
	PathSize  uint32
}

type Sym struct { //ignore fieldalignment warning
	SymBase
	Path []byte
}

type ExtSymBase struct {
	HardLinks uint32
	PathSize  uint32
}

type ExtSym struct { //ignore fieldalignment warning
	ExtSymBase
	Path       []byte
	XattrIndex uint32
}

type Device struct {
	HardLinks uint32
	DeviceNum uint32
}

type ExtDevice struct {
	Device
	XattrIndex uint32
}

type IPC struct {
	HardLinks uint32
}

type ExtIPC struct {
	HardLinks  uint32
	XattrIndex uint32
}
