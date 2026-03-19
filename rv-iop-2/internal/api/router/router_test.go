package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit_PanicsWhenDependenciesNil(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Init(nil) should panic")
	}()
	Init(nil)
	t.Error("Init(nil) should have panicked")
}
