package data

import "io"

type zeroReader struct {
	size uint32
}

func (z *zeroReader) Read(p []byte) (n int, err error) {
	if len(p) > int(z.size) {
		for i := uint32(0); i < z.size; i++ {
			p[i] = 0
		}
		n = int(z.size)
		err = io.EOF
		z.size = 0
		return
	}
	for i := range p {
		p[i] = 0
	}
	z.size -= uint32(len(p))
	n = len(p)
	return
}

func (z zeroReader) Close() error {
	return nil
}
