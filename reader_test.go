package squashfs

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	goappimage "github.com/CalebQ42/GoAppImage"
)

const (
	appimageURL  = "https://github.com/srevinsaju/Firefox-Appimage/releases/download/firefox-v84.0.r20201221152838/firefox-84.0.r20201221152838-x86_64.AppImage"
	appimageName = "firefox-84.0.r20201221152838-x86_64.AppImage"
	//TODO: Find better raw squashfs example. This one takes too long to extract
	sfsURL  = "http://mirror.rackspace.com/archlinux/iso/2021.07.01/arch/x86_64/airootfs.sfs"
	sfsName = "airootfs.sfs"
)

func TestSquashfs(t *testing.T) {
	t.Parallel()
	sfsFil, err := os.Open("testing/" + sfsName)
	if os.IsNotExist(err) {
		err = downloadTestSfs("testing")
		if err != nil {
			t.Fatal(err)
		}
		sfsFil, err = os.Open("testing/" + appimageName)
		if err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	defer sfsFil.Close()
	rdr, err := NewReader(sfsFil)
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll("testing/" + sfsName + ".d")
	err = rdr.ExtractTo("testing/" + sfsName + ".d")
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal("No Problems")
}

func TestAppImage(t *testing.T) {
	t.Parallel()
	aiFil, err := os.Open("testing/" + appimageName)
	if os.IsNotExist(err) {
		err = downloadTestAppImage("testing")
		if err != nil {
			t.Fatal(err)
		}
		aiFil, err = os.Open("testing/" + appimageName)
		if err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	defer aiFil.Close()
	stat, _ := aiFil.Stat()
	ai := goappimage.NewAppImage("testing/" + appimageName)
	rdr, err := NewReader(io.NewSectionReader(aiFil, ai.Offset, stat.Size()-ai.Offset))
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll("testing/firefox")
	fil, err := rdr.Open("updater.ini")
	if err != nil {
		t.Fatal(err)
	}
	err = fil.(*File).ExtractTo("testing/firefox")
	// err = rdr.ExtractTo(wd + "/testing/firefox")
	t.Fatal(err)
}

func BenchmarkDragRace(b *testing.B) {
	aiFil, err := os.Open("testing/" + appimageName)
	if os.IsNotExist(err) {
		err = downloadTestAppImage("testing")
		if err != nil {
			b.Fatal(err)
		}
		aiFil, err = os.Open("testing/" + appimageName)
		if err != nil {
			b.Fatal(err)
		}
	} else if err != nil {
		b.Fatal(err)
	}
	stat, _ := aiFil.Stat()
	ai := goappimage.NewAppImage("testing/" + appimageName)
	os.RemoveAll("testing/unsquashFirefox")
	os.RemoveAll("testing/firefox")
	cmd := exec.Command("unsquashfs", "-d", "testing/unsquashFirefox", "-o", strconv.Itoa(int(ai.Offset)), aiFil.Name())
	start := time.Now()
	err = cmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	unsquashTime := time.Since(start)
	start = time.Now()
	rdr, err := NewReader(io.NewSectionReader(aiFil, ai.Offset, stat.Size()-ai.Offset))
	if err != nil {
		b.Fatal(err)
	}
	err = rdr.ExtractTo("testing/firefox")
	if err != nil {
		b.Fatal(err)
	}
	libTime := time.Since(start)
	b.Log("Unsqushfs:", unsquashTime.Round(time.Millisecond))
	b.Log("Library:", libTime.Round(time.Millisecond))
	b.Log("unsquashfs is", strconv.FormatFloat(float64(libTime.Milliseconds())/float64(unsquashTime.Milliseconds()), 'f', 2, 64)+"x faster")
	// b.Error("STOP ALREADY!")
}

func downloadTestSfs(dir string) error {
	//seems to time out on slow connections. Might fix that at some point... or not. It's just a test any...
	os.Mkdir(dir, os.ModePerm)
	sfs, err := os.Create(dir + "/" + sfsName)
	if err != nil {
		return err
	}
	defer sfs.Close()
	check := http.Client{
		CheckRedirect: func(r *http.Request, _ []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := check.Get(sfsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(sfs, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func downloadTestAppImage(dir string) error {
	//seems to time out on slow connections. Might fix that at some point... or not. It's just a test any...
	os.Mkdir(dir, os.ModePerm)
	appImage, err := os.Create(dir + "/" + appimageName)
	if err != nil {
		return err
	}
	defer appImage.Close()
	check := http.Client{
		CheckRedirect: func(r *http.Request, _ []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := check.Get(appimageURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(appImage, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
