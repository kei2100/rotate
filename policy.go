package rotate

// FileState holds the state of the current write destination file
type FileState struct {
	// openedAt Unix time
	OpenedAt int64
	// file size (when opened) + written bytes
	Size int64
}

// PolicyFunc is a type of rotate policy function
type PolicyFunc func(fileState FileState) bool

// NeedRotate reports whether need rotate
func (f PolicyFunc) NeedRotate(fileState FileState) bool {
	return f(fileState)
}

// SizeBasedPolicy returns size based rotate policy
func SizeBasedPolicy(size int64) PolicyFunc {
	return func(fileState FileState) bool {
		return fileState.Size >= size
	}
}

// TimeBasedPolicy returns time based rotate policy
func TimeBasedPolicy(fn func(openedAtUnix int64) bool) PolicyFunc {
	return func(fileState FileState) bool {
		return fn(fileState.OpenedAt)
	}
}
