package game

import "github.com/google/uuid"

type UUIDGenerator interface {
	Generate() string
}

type DefaultUUIDGenerator struct{}

func (g *DefaultUUIDGenerator) Generate() string {
	return uuid.New().String()
}

type MockUUIDGenerator struct {
	values []string
	index  int
}

func NewMockUUIDGenerator(values []string) *MockUUIDGenerator {
	return &MockUUIDGenerator{values: values, index: 0}
}

func (g *MockUUIDGenerator) Generate() string {
	if g.index >= len(g.values) {
		return ""
	}
	val := g.values[g.index]
	g.index++
	return val
}
