package dependency

import (
	"testing"

	"github.com/Meesho/rv-iop/internal/api/controller"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var _ controller.RecentlyViewedFeedImpl = (*mockRecentlyViewedFeedImpl)(nil)
var _ controller.RvInteractionsImpl = (*mockRvInteractionsImpl)(nil)

// mockRecentlyViewedFeedImpl implements controller.RecentlyViewedFeedImpl for tests.
type mockRecentlyViewedFeedImpl struct{}

func (m *mockRecentlyViewedFeedImpl) FetchRvSubstituteExploitFeed(c *gin.Context) {}
func (m *mockRecentlyViewedFeedImpl) FetchRvSubstituteExploreFeed(c *gin.Context) {}
func (m *mockRecentlyViewedFeedImpl) FetchRvSubstituteAdFeed(c *gin.Context)      {}

// mockRvInteractionsImpl implements controller.RvInteractionsImpl for tests.
type mockRvInteractionsImpl struct{}

func (m *mockRvInteractionsImpl) FetchRvInteractionsFeed(c *gin.Context)    {}
func (m *mockRvInteractionsImpl) FetchRvInteractionsWidget(c *gin.Context) {}

func TestNewDependencies_ReturnsStructWithGivenHandlers(t *testing.T) {
	rvFeed := &mockRecentlyViewedFeedImpl{}
	rvInteractions := &mockRvInteractionsImpl{}

	deps := NewDependencies(rvFeed, rvInteractions)

	assert.NotNil(t, deps)
	assert.Equal(t, rvFeed, deps.RecentlyViewedFeedImpl)
	assert.Equal(t, rvInteractions, deps.RvInteractionsImpl)
}

func TestNewDependencies_AllowsNilRvInteractionsImpl(t *testing.T) {
	rvFeed := &mockRecentlyViewedFeedImpl{}

	deps := NewDependencies(rvFeed, nil)

	assert.NotNil(t, deps)
	assert.Equal(t, rvFeed, deps.RecentlyViewedFeedImpl)
	assert.Nil(t, deps.RvInteractionsImpl)
}

func TestNewDependencies_AllowsNilRecentlyViewedFeedImpl(t *testing.T) {
	rvInteractions := &mockRvInteractionsImpl{}

	deps := NewDependencies(nil, rvInteractions)

	assert.NotNil(t, deps)
	assert.Nil(t, deps.RecentlyViewedFeedImpl)
	assert.Equal(t, rvInteractions, deps.RvInteractionsImpl)
}
