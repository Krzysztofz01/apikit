package utils

func IsDistinct[T comparable](s []T) bool {
	if s == nil {
		panic("utils: provided slice is nil")
	}

	m := make(map[T]bool, len(s))
	for _, entry := range s {
		if _, ok := m[entry]; ok {
			return false
		} else {
			m[entry] = true
		}
	}

	return true
}
