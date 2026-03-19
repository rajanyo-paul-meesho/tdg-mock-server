package dependency

import "github.com/Meesho/rv-iop/internal/api/controller"

type Dependencies struct {
	RecentlyViewedFeedImpl controller.RecentlyViewedFeedImpl
	RvInteractionsImpl     controller.RvInteractionsImpl // nil when deployable is deprecated (e.g. rviopweb)
}

func NewDependencies(
	RecentlyViewedFeedImpl controller.RecentlyViewedFeedImpl,
	RvInteractionsImpl controller.RvInteractionsImpl) *Dependencies {
	return &Dependencies{
		RecentlyViewedFeedImpl: RecentlyViewedFeedImpl,
		RvInteractionsImpl:     RvInteractionsImpl,
	}
}
