package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDistinctShouldReturnCorrectValue(t *testing.T) {
	cases := []struct {
		a        []int
		expected bool
	}{
		{a: []int{1, 2, 3}, expected: true},
		{a: []int{1, 2, 3, 3}, expected: false},
		{a: []int{}, expected: true},
	}

	for _, c := range cases {
		actual := IsDistinct(c.a)

		assert.Equal(t, c.expected, actual)
	}
}
