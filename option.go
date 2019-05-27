package rotate

import (
	"os"
)

type option struct {
	permission os.FileMode
	keeps      int
	policy     PolicyFunc
}

// OptionFunc let you change follow.Reader behavior.
type OptionFunc func(o *option)

// Default values
const (
	DefaultPermission = 0600
	DefaultKeeps      = 5
	DefaultSize       = 1024 * 1024 * 10
)

func (o *option) apply(opts ...OptionFunc) {
	o.permission = DefaultPermission
	o.keeps = DefaultKeeps
	o.policy = SizeBasedPolicy(DefaultSize)
	for _, fn := range opts {
		fn(o)
	}
}

// WithPermission let you change the file permission
func WithPermission(v os.FileMode) OptionFunc {
	return func(o *option) {
		o.permission = v
	}
}

// WithKeeps let you change keep count of the rotated files
func WithKeeps(v int) OptionFunc {
	return func(o *option) {
		o.keeps = v
	}
}

// WithSizeBasedPolicy let you change the rotate policy
func WithSizeBasedPolicy(size int64) OptionFunc {
	return func(o *option) {
		o.policy = SizeBasedPolicy(size)
	}
}

// WithTimeBasedPolicy let you change the rotate policy
func WithTimeBasedPolicy(fn func(openedAtUnix int64) bool) OptionFunc {
	return func(o *option) {
		o.policy = TimeBasedPolicy(fn)
	}
}
