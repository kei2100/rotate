// +build linux freebsd darwin

package file

import "os"

// OpenFile opens the named file
func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}
