package rotate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriter_Rotate(t *testing.T) {
	t.Parallel()

	dir := createTmpDir()
	defer dir.removeAll()

	nBytes := 100
	keeps := 2

	w, err := NewWriter(string(dir), "test.log", WithKeeps(keeps), WithConfigFunc(SizeBasedConfig(int64(nBytes))))
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := writeNCount(w, "a", nBytes); err != nil {
		t.Fatal(err)
	}
	if err := dir.waitFileCreated(time.Second, "test.log", "test.log.1"); err != nil {
		t.Fatal(err)
	}

	if err := writeNCount(w, "b", nBytes); err != nil {
		t.Fatal(err)
	}
	if err := dir.waitFileCreated(time.Second, "test.log", "test.log.1", "test.log.2"); err != nil {
		t.Fatal(err)
	}

	if err := writeNCount(w, "c", nBytes-1); err != nil {
		t.Fatal(err)
	}
	if err := dir.waitFileCreated(time.Second, "test.log", "test.log.1", "test.log.2"); err != nil {
		t.Fatal(err)
	}

	if err := containsNCount("c", nBytes-1, dir, "test.log"); err != nil {
		t.Fatal(err)
	}
	if err := containsNCount("b", nBytes, dir, "test.log.1"); err != nil {
		t.Fatal(err)
	}
	if err := containsNCount("a", nBytes, dir, "test.log.2"); err != nil {
		t.Fatal(err)
	}

	if err := writeNCount(w, "c", 1); err != nil {
		t.Fatal(err)
	}
	if err := dir.waitFileCreated(time.Second, "test.log", "test.log.1", "test.log.2"); err != nil {
		t.Fatal(err)
	}

	if err := dir.waitFileNotCreated(time.Second, "test.log.3"); err != nil {
		t.Fatal(err)
	}

	if err := containsNCount("c", nBytes, dir, "test.log.1"); err != nil {
		t.Fatal(err)
	}
	if err := containsNCount("b", nBytes, dir, "test.log.2"); err != nil {
		t.Fatal(err)
	}
	if err := emptyFile(dir, "test.log"); err != nil {
		t.Fatal(err)
	}
}

// TODO multigroutine

type tmpDir string

func createTmpDir() tmpDir {
	s, err := ioutil.TempDir("", "rotate-test")
	if err != nil {
		panic(err)
	}
	return tmpDir(s)
}

func (d tmpDir) removeAll() {
	os.RemoveAll(string(d))
}

func (d tmpDir) waitFileCreated(timeout time.Duration, filenames ...string) error {
	fn := func() error {
		fis, ioerr := ioutil.ReadDir(string(d))
		if ioerr != nil {
			return ioerr
		}
		var err error
		for _, fn := range filenames {
			var exists bool
			for _, fi := range fis {
				if fi.Name() == fn {
					exists = true
					break
				}
			}
			if !exists {
				if err == nil {
					err = fmt.Errorf("%s is not exists", fn)
				} else {
					err = fmt.Errorf("%s is not exists: %v", fn, err)
				}
			}
		}
		return err
	}
	return retry(timeout, 10*time.Millisecond, fn)
}

func (d tmpDir) waitFileNotCreated(wait time.Duration, filenames ...string) error {
	wa := time.After(wait)
	<-wa

	fis, ioerr := ioutil.ReadDir(string(d))
	if ioerr != nil {
		return ioerr
	}
	var err error
	for _, fn := range filenames {
		for _, fi := range fis {
			if fi.Name() == fn {
				if err == nil {
					err = fmt.Errorf("%s is exists", fn)
				} else {
					err = fmt.Errorf("%s is exists: %v", fn, err)
				}
				break
			}
		}
	}
	return err
}

func retry(timeout, interval time.Duration, fn func() error) error {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	to := time.After(timeout)

	for {
		err := fn()
		if err == nil {
			return nil
		}
		select {
		case <-tick.C:
			continue
		case <-to:
			return fmt.Errorf("timeout: %v", err)
		}
	}
}

func writeNCount(w io.Writer, s string, nCount int) error {
	for nCount > 0 {
		n, err := fmt.Fprint(w, s)
		if err != nil {
			return err
		}
		nCount -= n
	}
	return nil
}

func containsNCount(s string, nCount int, d tmpDir, filenames ...string) error {
	var buf bytes.Buffer
	for _, fn := range filenames {
		b, err := ioutil.ReadFile(filepath.Join(string(d), fn))
		if err != nil {
			return err
		}
		buf.Write(b)
	}
	cat := buf.String()
	count := strings.Count(cat, s)
	if count != nCount {
		return fmt.Errorf("%s contains %d count %v", s, count, filenames)
	}
	return nil
}

func emptyFile(d tmpDir, filename string) error {
	fi, err := os.Stat(filepath.Join(string(d), filename))
	if err != nil {
		return err
	}
	if fi.Size() > 0 {
		return fmt.Errorf("%s is not empty (%d bytes)", filename, fi.Size())
	}
	return nil
}
