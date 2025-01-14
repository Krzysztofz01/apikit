package utils

type Set[T comparable] interface {
	Add(element T) bool
	Remove(element T) bool
	Contains(element T) bool
	Clear()
}

type set[T comparable] struct {
	data map[T]bool
}

func (s *set[T]) Add(element T) bool {
	if _, ok := s.data[element]; ok {
		return false
	}

	s.data[element] = true
	return true
}

func (s *set[T]) Clear() {
	s.data = make(map[T]bool, 100)
}

func (s *set[T]) Contains(element T) bool {
	_, ok := s.data[element]
	return ok
}

func (s *set[T]) Remove(element T) bool {
	if _, ok := s.data[element]; !ok {
		return false
	}

	delete(s.data, element)
	return true
}

func NewEmptySet[T comparable]() Set[T] {
	return &set[T]{
		data: make(map[T]bool, 100),
	}
}
