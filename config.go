package rotate

type ConfigFunc func(currentState FileState) bool

func (f ConfigFunc) NeedRotate(currentState FileState) bool {
	return f(currentState)
}

func SizeConfig(size int64) ConfigFunc {
	return func(currentState FileState) bool {
		return currentState.Size() > size
	}
}
