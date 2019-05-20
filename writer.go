package rotate

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kei2100/rotate/logger"
)

func NewWriter(dir, filename string, opts ...OptionFunc) (*Writer, error) {
	var opt option
	opt.apply(opts...)

	filePath := filepath.Join(dir, filename)
	f, err := openFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, opt.permission)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &Writer{
		f: f,
		state: state{
			FileState: FileState{openedAt: time.Now().Unix(), size: fi.Size()},
		},
		filePath: filePath,
		opt:      opt,
	}, nil
}

type Writer struct {
	mu       sync.RWMutex
	f        *os.File
	state    state
	filePath string
	opt      option
}

func (w *Writer) Write(p []byte) (int, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	n, err := w.f.Write(p)
	if err != nil {
		return n, err
	}
	w.state.addSize(int64(n))
	if !w.opt.conf.NeedRotate(w.state.FileState) {
		return n, nil
	}
	if !w.state.compareAndSwapAsRotating() {
		return n, nil
	}

	go func(current *os.File, st *state, opt option) {
		nextTmpPath, err := randPath(w.filePath)
		if err != nil {
			logger.Println(err)
			st.compareAndSwapAsNotRotating()
			return
		}
		next, err := openFile(nextTmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, opt.permission)
		if err != nil {
			logger.Println(err)
			st.compareAndSwapAsNotRotating()
			return
		}
		if err := pushAndShiftKeeps(w.filePath, opt.keeps); err != nil {
			logger.Println(err)
			st.compareAndSwapAsNotRotating()
			return
		}
		if err := os.Rename(nextTmpPath, w.filePath); err != nil {
			logger.Printf("rotate: failed to rename %s to %s: %+v", nextTmpPath, w.filePath, err)
			st.compareAndSwapAsNotRotating()
			return
		}

		w.mu.Lock()
		defer w.mu.Unlock()

		if st.isClosed() {
			if err := next.Close(); err != nil {
				logger.Printf("rotate: an error occurred while closing next file: %+v", err)
			}
			return
		}
		if err := current.Close(); err != nil {
			logger.Printf("rotate: an error occurred while closing current file: %+v", err)
			// not return
		}
		w.f = next
		w.state = state{
			FileState: FileState{openedAt: time.Now().Unix()},
		}
	}(w.f, &w.state, w.opt)

	return n, nil
}

func (w *Writer) Close() error {
	w.mu.RLock()
	w.state.storeAsClosed()
	err := w.f.Close()
	w.mu.RUnlock()

	return err
}

func formatRotatedPath(currentPath string, num int) string {
	return fmt.Sprintf("%s.%d", currentPath, num)
}

func randPath(currentFilePath string) (string, error) {
	pid := os.Getpid()
	nano := time.Now().UnixNano()
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("rotate: failed to generate rand path: %+v", err)
	}
	return fmt.Sprintf("%s-%d-%d-%x", currentFilePath, pid, nano, b), nil
}

//   e.g. currentPath "log", keeps 3
//   - log > log.1 | log.1 > log.2 | log.2 > log.3 | log.3 > remove
//   - log > log.1 | log.1 > log.2 |               | log.3 > not change
func pushAndShiftKeeps(filePath string, keeps int) error {
	files := make([]string, 0, keeps+1)
	for i := keeps; i > 0; i-- {
		p := formatRotatedPath(filePath, i)
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("rotate: failed to get stat %s: %+v", p, err)
		}
		files = append(files, p)
	}
	// - [log.3 log.2 log.1 log]
	// - [log.3 log.1 log]
	files = append(files, filePath)

	if len(files) > keeps {
		if err := os.Remove(files[0]); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("rotate: failed to remove %s", files[0])
		}
		// [log.2 log.1 log]
		files = files[1:]
	}
	for i, old := range files {
		nw := formatRotatedPath(filePath, len(files)-i)
		if err := os.Rename(old, nw); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("rotate: failed to rename %s to %s", old, nw)
		}
	}
	return nil
}
