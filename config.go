package rotate

import "time"

// ConfigFunc is a type of rotate configuration function
type ConfigFunc func(fileState FileState) bool

// NeedRotate reports whether need rotate
func (f ConfigFunc) NeedRotate(fileState FileState) bool {
	return f(fileState)
}

// SizeBasedConfig returns size based rotate configuration
func SizeBasedConfig(size int64) ConfigFunc {
	return func(fileState FileState) bool {
		return fileState.Size() > size
	}
}

// TimeBasedConfig returns time based rotate configuration
func TimeBasedConfig(elapsed time.Duration) ConfigFunc {
	return func(fileState FileState) bool {
		return time.Now().Unix()-fileState.openedAt >= int64(elapsed)
	}
}
