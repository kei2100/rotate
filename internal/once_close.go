package internal

import (
	"os"
	"sync"
)

type OnceCloseFile struct {
	once sync.Once
	*os.File
}

func (ocf *OnceCloseFile) Close() error {
	var err error
	ocf.once.Do(func() {
		err = ocf.File.Close()
	})
	return err
}
