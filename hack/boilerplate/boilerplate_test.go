package main

import (
	"github.com/magiconair/properties/assert"

	"testing"
)

func TestPassContentThesame(t *testing.T) {
	t1 := []string{"this is a just a string"}
	t2 := []string{"this is a just a string"}

	assert.Equal(t, IsContentTheSame(t1, t2), true)
}

func TestFailContentThesame(t *testing.T) {
	t1 := []string{"this is a just a string1"}
	t2 := []string{"this is a just a string2"}

	assert.Equal(t, !IsContentTheSame(t1, t2), true)
}
