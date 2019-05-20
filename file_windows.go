package rotate

import (
	"os"

	"github.com/kei2100/filesharedelete"
)

func openFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return filesharedelete.OpenFile(name, flag, perm)
}
