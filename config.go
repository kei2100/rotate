package rotate

import "time"

// FileState holds the state of the current write destination file
type FileState struct {
	// openedAt Unix time
	OpenedAt int64
	// file size (when opened) + written bytes
	Size int64
}

// ConfigFunc is a type of rotate configuration function
type ConfigFunc func(fileState FileState) bool

// NeedRotate reports whether need rotate
func (f ConfigFunc) NeedRotate(fileState FileState) bool {
	return f(fileState)
}

// SizeBasedConfig returns size based rotate configuration
func SizeBasedConfig(size int64) ConfigFunc {
	return func(fileState FileState) bool {
		return fileState.Size >= size
	}
}

// TimeBasedConfig returns time based rotate configuration
func TimeBasedConfig(elapsed time.Duration) ConfigFunc {
	return func(fileState FileState) bool {
		return time.Now().Unix()-fileState.OpenedAt >= int64(elapsed)
	}
}
