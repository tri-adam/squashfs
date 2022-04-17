package squashfs

import (
	"github.com/jacobsa/fuse/fuseops"
)

type handler struct {
	context fuseops.OpContext
}
