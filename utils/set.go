package utils

type Set struct {
	mapForSet map[string]struct{}
}

func (s *Set) Len() int { return len(s.mapForSet) }

func NewSet() *Set {
	s := &Set{}
	s.mapForSet = make(map[string]struct{})
	return s
}

func (s *Set) Add(value string) {
	s.mapForSet[value] = struct{}{}
}

func (s *Set) Remove(value string) {
	delete(s.mapForSet, value)
}

func (s *Set) Contains(value string) bool {
	_, c := s.mapForSet[value]
	return c
}
