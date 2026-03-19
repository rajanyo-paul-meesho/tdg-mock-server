package router

import (
	"github.com/Meesho/go-core/httpframework"
	"github.com/Meesho/rv-iop/internal/api/controller"
	"github.com/Meesho/rv-iop/internal/dependency"
	"github.com/rs/zerolog/log"
)

// Init expects http framework to be initialized before calling this function
func Init(init *dependency.Dependencies) {
	if init == nil {
		log.Panic().Msg("dependencies are empty")
	}

	api := httpframework.Instance().Group("/api")
	{
		v1 := api.Group("/v1")
		rvFeed := v1.Group("/entities")
		{
			rvSubstituteFeed := rvFeed.Group("/rv-substitute")
			{
				rvSubstituteFeed.POST("/ad", init.RecentlyViewedFeedImpl.FetchRvSubstituteAdFeed)
				rvSubstituteFeed.POST("/exploit", init.RecentlyViewedFeedImpl.FetchRvSubstituteExploitFeed)
				rvSubstituteFeed.POST("/explore", init.RecentlyViewedFeedImpl.FetchRvSubstituteExploreFeed)
			}
		}
		// Recently viewed endpoints (only when RV interactions impl is present; rviopweb deployable is deprecated)
		if init.RvInteractionsImpl != nil {
			recentlyViewed := v1.Group("/recently-viewed")
			{
				recentlyViewed.POST("/feed", init.RvInteractionsImpl.FetchRvInteractionsFeed)
				recentlyViewed.POST("/widget", init.RvInteractionsImpl.FetchRvInteractionsWidget)
			}
		}
	}
	// Init health check
	httpframework.Instance().GET("/health", controller.Health)
}
