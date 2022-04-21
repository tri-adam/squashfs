package squashfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	goappimage "github.com/CalebQ42/GoAppImage"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

const (
	appImageURL  = "https://github.com/srevinsaju/Firefox-Appimage/releases/download/firefox-v84.0.r20201221152838/firefox-84.0.r20201221152838-x86_64.AppImage"
	appImageName = "firefox-84.0.r20201221152838-x86_64.AppImage"
	squashfsURL  = "https://darkstorm.tech/LinuxPATest.sfs"
	squashfsName = "out.sfs"
)

func TestSquashfs(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	squashFil, err := os.Open(wd + "/testing/" + squashfsName)
	if os.IsNotExist(err) {
		err = downloadTestSquash(wd + "/testing")
		if err != nil {
			t.Fatal(err)
		}
		squashFil, err = os.Open(wd + "/testing/" + squashfsName)
	}
	if err != nil {
		t.Fatal(err)
	}
	rdr, err := NewSquashfsReader(squashFil)
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(wd + "/testing/" + squashfsName + ".d")
	op := DefaultOptions()
	op.Verbose = true
	err = rdr.ExtractWithOptions(wd+"/testing/"+squashfsName+".d", op)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal("No Problems")
}

func TestSquashfsFromReader(t *testing.T) {
	resp, err := http.DefaultClient.Get(squashfsURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	rdr, err := NewSquashfsReaderFromReader(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll("testing/" + squashfsName + ".d")
	op := DefaultOptions()
	op.Verbose = true
	err = rdr.ExtractWithOptions("testing/"+squashfsName+".d", op)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal("No Problems")
}

func TestAppImage(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	aiFil, err := os.Open(wd + "/testing/" + appImageName)
	if os.IsNotExist(err) {
		err = downloadTestAppImage(wd + "/testing")
		if err != nil {
			t.Fatal(err)
		}
		aiFil, err = os.Open(wd + "/testing/" + appImageName)
		if err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	defer aiFil.Close()
	stat, _ := aiFil.Stat()
	ai := goappimage.NewAppImage(wd + "/testing/" + appImageName)
	rdr, err := NewReader(io.NewSectionReader(aiFil, ai.Offset, stat.Size()-ai.Offset))
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(wd + "/testing/firefox")
	err = rdr.ExtractTo(wd + "/testing/firefox")
	t.Fatal(err)
}

func TestUnsquashfs(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	aiFil, err := os.Open(wd + "/testing/" + appImageName)
	if os.IsNotExist(err) {
		err = downloadTestAppImage(wd + "/testing")
		if err != nil {
			t.Fatal(err)
		}
		aiFil, err = os.Open(wd + "/testing/" + appImageName)
		if err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(wd + "/testing/unsquashFirefox")
	os.RemoveAll(wd + "/testing/firefox")
	ai := goappimage.NewAppImage(wd + "/testing/" + appImageName)
	fmt.Println("Command:", "unsquashfs", "-d", wd+"/testing/unsquashFirefox", "-o", strconv.Itoa(int(ai.Offset)), aiFil.Name())
	cmd := exec.Command("unsquashfs", "-d", wd+"/testing/unsquashFirefox", "-o", strconv.Itoa(int(ai.Offset)), aiFil.Name())
	start := time.Now()
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(time.Since(start))
	t.Fatal("HI")
}

func BenchmarkAppImageDragRace(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	aiFil, err := os.Open(wd + "/testing/" + appImageName)
	if os.IsNotExist(err) {
		err = downloadTestAppImage(wd + "/testing")
		if err != nil {
			b.Fatal(err)
		}
		aiFil, err = os.Open(wd + "/testing/" + appImageName)
		if err != nil {
			b.Fatal(err)
		}
	} else if err != nil {
		b.Fatal(err)
	}
	stat, _ := aiFil.Stat()
	ai := goappimage.NewAppImage(wd + "/testing/" + appImageName)
	os.RemoveAll(wd + "/testing/unsquashFirefox")
	os.RemoveAll(wd + "/testing/firefox")
	cmd := exec.Command("unsquashfs", "-d", wd+"/testing/unsquashFirefox", "-o", strconv.Itoa(int(ai.Offset)), aiFil.Name())
	start := time.Now()
	err = cmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	unsquashTime := time.Since(start)
	start = time.Now()
	rdr, err := NewSquashfsReader(io.NewSectionReader(aiFil, ai.Offset, stat.Size()-ai.Offset))
	if err != nil {
		b.Fatal(err)
	}
	err = rdr.ExtractTo(wd + "/testing/firefox")
	if err != nil {
		b.Fatal(err)
	}
	libTime := time.Since(start)
	b.Log("Unsqushfs:", unsquashTime.Round(time.Millisecond))
	b.Log("Library:", libTime.Round(time.Millisecond))
	b.Log("unsquashfs is", strconv.FormatFloat(float64(libTime.Milliseconds())/float64(unsquashTime.Milliseconds()), 'f', 2, 64)+"x faster")
	b.Error("STOP ALREADY!")
}

func BenchmarkSquashfsDragRace(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	aiFil, err := os.Open(wd + "/testing/" + squashfsName)
	if os.IsNotExist(err) {
		b.Fatal(err)
	} else if err != nil {
		b.Fatal(err)
	}
	os.RemoveAll(wd + "/testing/unsquashSquash")
	os.RemoveAll(wd + "/testing/squash")
	cmd := exec.Command("unsquashfs", "-d", wd+"/testing/unsquashSquash", aiFil.Name())
	start := time.Now()
	err = cmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	unsquashTime := time.Since(start)
	start = time.Now()
	rdr, err := NewSquashfsReader(aiFil)
	if err != nil {
		b.Fatal(err)
	}
	err = rdr.ExtractTo(wd + "/testing/squash")
	if err != nil {
		b.Fatal(err)
	}
	libTime := time.Since(start)
	b.Log("Unsqushfs:", unsquashTime.Round(time.Millisecond))
	b.Log("Library:", libTime.Round(time.Millisecond))
	b.Log("unsquashfs is", strconv.FormatFloat(float64(libTime.Milliseconds())/float64(unsquashTime.Milliseconds()), 'f', 2, 64)+"x faster")
	b.Error("STOP ALREADY!")
}

func BenchmarkSingleFile(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	aiFil, err := os.Open(wd + "/testing/" + appImageName)
	if os.IsNotExist(err) {
		err = downloadTestAppImage(wd + "/testing")
		if err != nil {
			b.Fatal(err)
		}
		aiFil, err = os.Open(wd + "/testing/" + appImageName)
		if err != nil {
			b.Fatal(err)
		}
	} else if err != nil {
		b.Fatal(err)
	}
	stat, _ := aiFil.Stat()
	ai := goappimage.NewAppImage(wd + "/testing/" + appImageName)
	os.RemoveAll(wd + "/testing/omni.ja.unsquash")
	os.RemoveAll(wd + "/testing/omni.ja.squashfs")
	cmd := exec.Command("unsquashfs", "-d", wd+"/testing/omni.ja.unsquash", "-o", strconv.Itoa(int(ai.Offset)), aiFil.Name(), "omni.ja")
	start := time.Now()
	err = cmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	unsquashTime := time.Since(start)
	outFil, err := os.Create(wd + "/testing/omni.ja.squashfs")
	if err != nil {
		b.Fatal(err)
	}
	start = time.Now()
	rdr, err := NewSquashfsReader(io.NewSectionReader(aiFil, ai.Offset, stat.Size()-ai.Offset))
	if err != nil {
		b.Fatal(err)
	}
	fil, err := rdr.Open("omni.ja")
	if err != nil {
		b.Fatal(err)
	}
	_, err = io.Copy(outFil, fil)
	if err != nil {
		b.Fatal(err)
	}
	libTime := time.Since(start)
	b.Log("Unsqushfs:", unsquashTime.Round(time.Millisecond))
	b.Log("Library:", libTime.Round(time.Millisecond))
	b.Log("unsquashfs is", strconv.FormatFloat(float64(libTime.Milliseconds())/float64(unsquashTime.Milliseconds()), 'f', 2, 64)+"x faster")
	b.Error("STOP ALREADY!")
}

func BenchmarkSTUFF(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	_, err = os.Open(wd + "/testing/" + appImageName)
	if os.IsNotExist(err) {
		err = downloadTestAppImage(wd + "/testing")
		if err != nil {
			b.Fatal(err)
		}
		_, err = os.Open(wd + "/testing/" + appImageName)
		if err != nil {
			b.Fatal(err)
		}
	} else if err != nil {
		b.Fatal(err)
	}
	//Compress the appimage first to create a test file.
	var zstdFile, gzipFile *os.File
	if zstdFile, err = os.Open(wd + "/testing/" + appImageName + ".zst"); err != nil {
		cmd := exec.Command("zstd", wd+"/testing/"+appImageName, "-o", wd+"/testing/"+appImageName+".zst")
		cmd.Run()
		if zstdFile, err = os.Open(wd + "/testing/" + appImageName + ".zst"); err != nil {
			b.Fatal(err)
		}
	}
	if gzipFile, err = os.Open(wd + "/testing/" + appImageName + ".gz"); err != nil {
		cmd := exec.Command("zstd", wd+"/testing/"+appImageName, "-o", wd+"/testing/"+appImageName+".gz", "--format=gzip")
		cmd.Run()
		if gzipFile, err = os.Open(wd + "/testing/" + appImageName + ".gz"); err != nil {
			b.Fatal(err)
		}
	}
	os.Remove(wd + "/testing/firefox.out")
	outFil, _ := os.Create(wd + "/testing/firefox.out")
	var zstdTime, gzipTime time.Duration
	timeStart := time.Now()
	zRdr, err := zstd.NewReader(zstdFile)
	if err != nil {
		b.Fatal(err)
	}
	_, err = io.Copy(outFil, zRdr)
	if err != nil {
		b.Fatal(err)
	}
	zstdTime = time.Since(timeStart)
	os.Remove(wd + "/testing/firefox.out")
	outFil, _ = os.Create(wd + "/testing/firefox.out")
	timeStart = time.Now()
	gRdr, err := gzip.NewReader(gzipFile)
	if err != nil {
		b.Fatal(err)
	}
	_, err = io.Copy(outFil, gRdr)
	if err != nil {
		b.Fatal(err)
	}
	gzipTime = time.Since(timeStart)
	b.Log("gzip Library:", gzipTime.Round(time.Millisecond))
	b.Log("zstd Library:", zstdTime.Round(time.Millisecond))
	os.Remove(wd + "/testing/firefox.out")
	cmd := exec.Command("zstd", "-d", wd+"/testing/"+appImageName+".gz", "--format=gzip", "-o", wd+"/testing/firefox.out")
	timeStart = time.Now()
	err = cmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	gzipTime = time.Since(timeStart)
	os.Remove(wd + "/testing/firefox.out")
	cmd = exec.Command("zstd", "-d", wd+"/testing/"+appImageName+".zst", "-o", wd+"/testing/firefox.out")
	timeStart = time.Now()
	err = cmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	zstdTime = time.Since(timeStart)
	b.Log("gzip cmd:", gzipTime.Round(time.Millisecond))
	b.Log("zstd cmd:", zstdTime.Round(time.Millisecond))
	b.Error("STOP ALREADY!")
}

func downloadTestAppImage(dir string) error {
	//seems to time out on slow connections. Might fix that at some point... or not. It's just a test...
	os.Mkdir(dir, os.ModePerm)
	appImage, err := os.Create(dir + "/" + appImageName)
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
	resp, err := check.Get(appImageURL)
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

func downloadTestSquash(dir string) error {
	//seems to time out on slow connections. Might fix that at some point... or not. It's just a test...
	os.Mkdir(dir, os.ModePerm)
	sfs, err := os.Create(dir + "/" + squashfsName)
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
	resp, err := check.Get(squashfsURL)
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

func TestCreateSquashFromAppImage(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir(wd+"/testing", 0777)
	if err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}
	_, err = os.Open(wd + "/testing/" + appImageName)
	if os.IsNotExist(err) {
		err = downloadTestAppImage(wd + "/testing")
		if err != nil {
			t.Fatal(err)
		}
		_, err = os.Open(wd + "/testing/" + appImageName)
		if err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	ai := goappimage.NewAppImage(wd + "/testing/" + appImageName)
	aiFil, err := os.Open(wd + "/testing/" + appImageName)
	if err != nil {
		t.Fatal(err)
	}
	defer aiFil.Close()
	aiFil.Seek(ai.Offset, 0)
	os.Remove(wd + "/testing/" + appImageName + ".squashfs")
	aiSquash, err := os.Create(wd + "/testing/" + appImageName + ".squashfs")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(aiSquash, aiFil)
	if err != nil {
		t.Fatal(err)
	}
}
