package controller

import (
	"testing"

	"github.com/Meesho/go-core/api"
	rvinteraction "github.com/Meesho/rv-iop/internal/api/handler/rv-interaction"
	"github.com/stretchr/testify/assert"
)

type mockRvInteractionsFeedImpl struct{}
type mockRvInteractionsWidgetImpl struct{}

func (m *mockRvInteractionsFeedImpl) FetchRvInteractionsFeed(_ *rvinteraction.GetRvInteractionsFeedRequest, _ *api.RequestContext) (*rvinteraction.GetRvInteractionsFeedResponse, *api.Error) {
	return nil, nil
}
func (m *mockRvInteractionsWidgetImpl) FetchRvInteractionsWidget(_ *rvinteraction.GetRvInteractionsWidgetRequest, _ *api.RequestContext) (*rvinteraction.GetRvInteractionsWidgetResponse, *api.Error) {
	return nil, nil
}

func TestNewStandardRvInteractionsImpl_ReturnsStructWhenAllDepsProvided(t *testing.T) {
	feedImpl := &mockRvInteractionsFeedImpl{}
	widgetImpl := &mockRvInteractionsWidgetImpl{}

	impl := NewStandardRvInteractionsImpl(feedImpl, widgetImpl)

	assert.NotNil(t, impl)
	assert.Equal(t, feedImpl, impl.RvInteractionsFeedHandler)
	assert.Equal(t, widgetImpl, impl.RvInteractionsWidgetHandler)
}

func TestNewStandardRvInteractionsImpl_PanicsWhenFeedHandlerNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r, "rv interactions feed handler cannot be nil")
		}
	}()
	NewStandardRvInteractionsImpl(nil, &mockRvInteractionsWidgetImpl{})
	t.Error("expected panic when feed handler is nil")
}

func TestNewStandardRvInteractionsImpl_PanicsWhenWidgetHandlerNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r, "rv interaction widget handler cannot be nil")
		}
	}()
	NewStandardRvInteractionsImpl(&mockRvInteractionsFeedImpl{}, nil)
	t.Error("expected panic when widget handler is nil")
}
