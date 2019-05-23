package file

import (
	"os"

	"github.com/kei2100/filesharedelete"
)

// OpenFile opens the named file
func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return filesharedelete.OpenFile(name, flag, perm)
}
