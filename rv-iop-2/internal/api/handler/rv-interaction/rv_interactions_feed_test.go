package rvinteraction

import (
	"testing"

	"github.com/Meesho/rv-iop/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewStandardRvInteractionsFeedImpl_ReturnsStructWhenAllDepsProvided(t *testing.T) {
	svc := &config.Service{}
	svc.App.Name = "test"
	iopHandler := &mockIopConfigHandler{}

	impl := NewStandardRvInteractionsFeedImpl(svc, iopHandler)

	assert.NotNil(t, impl)
	assert.Equal(t, svc, impl.ServiceConf)
	assert.Equal(t, iopHandler, impl.IopConfigHandler)
}

func TestNewStandardRvInteractionsFeedImpl_PanicsWhenServiceConfNil(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Contains(t, r, "service conf cannot be nil")
	}()
	NewStandardRvInteractionsFeedImpl(nil, &mockIopConfigHandler{})
	t.Error("expected panic when serviceConf is nil")
}

func TestNewStandardRvInteractionsFeedImpl_PanicsWhenIopConfigHandlerNil(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Contains(t, r, "iop config handler cannot be nil")
	}()
	NewStandardRvInteractionsFeedImpl(&config.Service{}, nil)
	t.Error("expected panic when iopConfigHandler is nil")
}
