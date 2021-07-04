package squashfs_test

import "testing"

func TestRandomStuff(t *testing.T) {
	hi := 1
	switch hi {
	case 0:
		println("HI")
	case 1:
		println("TODAY")
		fallthrough
	case 2:
		println("GoTACA")
	case 3:
		println("GIBBLE")
	case 4:
		println("HIIIE")
	}
	t.Error("OOPS")
}
