// +build linux freebsd darwin

package rotate

import "os"

func openFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}
