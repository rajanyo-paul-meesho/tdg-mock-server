package controller

import (
	"testing"

	"github.com/Meesho/go-core/api"
	"github.com/Meesho/rv-iop/internal/api/handler/ad"
	"github.com/Meesho/rv-iop/internal/api/handler/exploit"
	"github.com/Meesho/rv-iop/internal/api/handler/explore"
	"github.com/stretchr/testify/assert"
)

type mockExploitFeedImpl struct{}
type mockExploreFeedImpl struct{}
type mockAdFeedImpl struct{}

func (m *mockExploitFeedImpl) FetchRvSubstitute(_ *exploit.GetRvSubstituteOrganicFeedRequest, _ *api.RequestContext) (*exploit.GetRvSubstituteOrganicFeedResponse, *api.Error) {
	return nil, nil
}
func (m *mockExploreFeedImpl) FetchRvSubstituteCt(_ *explore.GetRvSubstituteCtFeedRequest, _ *api.RequestContext) (*explore.GetRvSubstituteCtFeedResponse, *api.Error) {
	return nil, nil
}
func (m *mockAdFeedImpl) FetchRvSubstituteAd(_ *ad.GetRvSubstituteAdFeedRequest, _ *api.RequestContext) (*ad.GetRvSubstituteAdFeedResponse, *api.Error) {
	return nil, nil
}

func TestNewStandardRecentlyViewedFeedImpl_ReturnsStructWhenAllDepsProvided(t *testing.T) {
	exploitImpl := &mockExploitFeedImpl{}
	exploreImpl := &mockExploreFeedImpl{}
	adImpl := &mockAdFeedImpl{}

	impl := NewStandardRecentlyViewedFeedImpl(exploitImpl, exploreImpl, adImpl)

	assert.NotNil(t, impl)
	assert.Equal(t, exploitImpl, impl.RvSubstituteFeedImpl)
	assert.Equal(t, exploreImpl, impl.RvSubstituteCtFeedImpl)
	assert.Equal(t, adImpl, impl.RvSubstituteAdFeedImpl)
}

func TestNewStandardRecentlyViewedFeedImpl_PanicsWhenExploitNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r, "exploit feed controller interface cannot be nil")
		}
	}()
	NewStandardRecentlyViewedFeedImpl(nil, &mockExploreFeedImpl{}, &mockAdFeedImpl{})
	t.Error("expected panic when exploit is nil")
}

func TestNewStandardRecentlyViewedFeedImpl_PanicsWhenExploreNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r, "explore feed controller interface cannot be nil")
		}
	}()
	NewStandardRecentlyViewedFeedImpl(&mockExploitFeedImpl{}, nil, &mockAdFeedImpl{})
	t.Error("expected panic when explore is nil")
}

func TestNewStandardRecentlyViewedFeedImpl_PanicsWhenAdNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r, "ad feed controller interface cannot be nil")
		}
	}()
	NewStandardRecentlyViewedFeedImpl(&mockExploitFeedImpl{}, &mockExploreFeedImpl{}, nil)
	t.Error("expected panic when ad is nil")
}
