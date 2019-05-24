package rotate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/go-ps"

	"golang.org/x/sync/errgroup"
)

func TestWriter_Rotate(t *testing.T) {
	t.Parallel()

	dir := createTmpDir()
	defer dir.removeAll()

	const nBytes = 100
	const keeps = 2

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

func TestWriter_Rotate_Parallel(t *testing.T) {
	t.Parallel()

	dir := createTmpDir()
	defer dir.removeAll()

	const nBytes = 100
	const keeps = 2
	const nGoroutines = 10

	w, err := NewWriter(string(dir), "test.log", WithKeeps(keeps), WithConfigFunc(SizeBasedConfig(int64(nBytes))))
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := nGroutinesDo(nGoroutines, func() error { return writeNCount(w, "a", 9) }); err != nil { // a * 90 bytes
		t.Fatal(err)
	}
	if err := dir.waitFileCreated(time.Second, "test.log"); err != nil {
		t.Fatal(err)
	}
	if err := dir.waitFileNotCreated(200*time.Millisecond, "test.log.1"); err != nil {
		t.Fatal(err)
	}

	if err := nGroutinesDo(nGoroutines, func() error { return writeNCount(w, "b", 8) }); err != nil { // b * 80 bytes
		t.Fatal(err)
	}
	if err := dir.waitFileCreated(time.Second, "test.log", "test.log.1"); err != nil {
		t.Fatal(err)
	}
	if err := dir.waitFileNotCreated(200*time.Millisecond, "test.log.2"); err != nil {
		t.Fatal(err)
	}

	if err := nGroutinesDo(nGoroutines, func() error { return writeNCount(w, "c", 11) }); err != nil { // c * 110 bytes
		t.Fatal(err)
	}
	if err := dir.waitFileCreated(time.Second, "test.log", "test.log.1", "test.log.2"); err != nil {
		t.Fatal(err)
	}
	if err := dir.waitFileNotCreated(200*time.Millisecond, "test.log.3"); err != nil {
		t.Fatal(err)
	}

	if err := containsNCount("a", 90, dir, "test.log.1", "test.log.2"); err != nil {
		t.Fatal(err)
	}
	if err := containsNCount("b", 80, dir, "test.log.1", "test.log.2"); err != nil {
		t.Fatal(err)
	}
	if err := containsNCount("c", 110, dir, "test.log", "test.log.1"); err != nil {
		t.Fatal(err)
	}
}

func TestWriter_Rotate_WhileOpeningFileFromAnotherProcess(t *testing.T) {
	t.Parallel()

	dir := createTmpDir()
	defer dir.removeAll()

	const nBytes = 100
	const keeps = 2

	w, err := NewWriter(string(dir), "test.log", WithKeeps(keeps), WithConfigFunc(SizeBasedConfig(int64(nBytes))))
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	writeNCount(w, "a", nBytes-1)

	prog := buildOpenFileProg(dir)
	cmd := exec.Command(prog, filepath.Join(string(dir), "test.log"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() { cmd.Wait() }()
	time.Sleep(500 * time.Millisecond)

	proc, err := ps.FindProcess(cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if proc == nil {
		t.Fatal("process not found")
	}

	writeNCount(w, "a", 1)
	time.Sleep(500 * time.Millisecond)

	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	writeNCount(w, "a", 1)
	if err := dir.waitFileCreated(time.Second, "test.log", "test.log.1"); err != nil {
		t.Fatal(err)
	}
	if err := containsNCount("a", nBytes+1, dir, "test.log", "test.log.1"); err != nil {
		t.Fatal(err)
	}
}

func Test_pushAndShiftKeeps(t *testing.T) {
	t.Parallel()

	tt := []struct {
		file            string
		rotatedFiles    []string
		keeps           int
		wantIncludes    []string
		wantNotIncludes []string
	}{
		{
			file:         "test.log",
			keeps:        3,
			wantIncludes: []string{"test.log.1"},
		},
		{
			file:         "test.log",
			rotatedFiles: []string{"test.log.1"},
			keeps:        3,
			wantIncludes: []string{"test.log.1", "test.log.2"},
		},
		{
			file:         "test.log",
			rotatedFiles: []string{"test.log.1", "test.log.2"},
			keeps:        3,
			wantIncludes: []string{"test.log.1", "test.log.2", "test.log.3"},
		},
		{
			file:            "test.log",
			rotatedFiles:    []string{"test.log.1", "test.log.2", "test.log.3"},
			keeps:           3,
			wantIncludes:    []string{"test.log.1", "test.log.2", "test.log.3"},
			wantNotIncludes: []string{"test.log.4"},
		},
		{
			file:            "test.log",
			rotatedFiles:    []string{"test.log.2", "test.log.3"},
			keeps:           3,
			wantIncludes:    []string{"test.log.1", "test.log.2", "test.log.3"},
			wantNotIncludes: []string{"test.log.4"},
		},
		{
			file:            "test.log",
			rotatedFiles:    []string{"test.log.1", "test.log.3"},
			keeps:           3,
			wantIncludes:    []string{"test.log.1", "test.log.2", "test.log.3"},
			wantNotIncludes: []string{"test.log.4"},
		},
	}
	for i, te := range tt {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			dir := createTmpDir()
			defer dir.removeAll()

			if err := touchFiles(dir, append(te.rotatedFiles, te.file)...); err != nil {
				t.Fatal(err)
			}
			if err := pushAndShiftKeeps(filepath.Join(string(dir), te.file), te.keeps); err != nil {
				t.Fatal(err)
			}
			if err := dir.waitFileCreated(time.Millisecond, te.wantIncludes...); err != nil {
				t.Error(err)
			}
			if err := dir.waitFileNotCreated(time.Millisecond, te.wantNotIncludes...); err != nil {
				t.Error(err)
			}
		})
	}
}

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

func touchFiles(d tmpDir, filenames ...string) error {
	for _, fn := range filenames {
		if err := ioutil.WriteFile(filepath.Join(string(d), fn), []byte{}, 0600); err != nil {
			return err
		}
	}
	return nil
}

func nGroutinesDo(n int, fn func() error) error {
	eg := errgroup.Group{}
	for i := 0; i < n; i++ {
		eg.Go(fn)
	}
	return eg.Wait()
}

func buildOpenFileProg(dir tmpDir) string {
	dstPath := filepath.Join(string(dir), "openfile"+binExtension())
	srcPath := filepath.Join("testdata", "cmd", "openfile", "main.go")
	cmd := exec.Command("go", "build", "-o", dstPath, srcPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	return dstPath
}

func binExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
