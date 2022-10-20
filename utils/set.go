package utils

var metricExists = struct{}{}

type Set struct {
	MapForSet map[string]struct{}
}

func NewSet() *Set {
	s := &Set{}
	s.MapForSet = make(map[string]struct{})
	return s
}

func (s *Set) Add(value string) {
	s.MapForSet[value] = metricExists
}

func (s *Set) Remove(value string) {
	delete(s.MapForSet, value)
}

func (s *Set) Contains(value string) bool {
	_, c := s.MapForSet[value]
	return c
}
