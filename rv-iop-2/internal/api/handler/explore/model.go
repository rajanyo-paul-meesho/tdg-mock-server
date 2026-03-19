package explore

import "github.com/Meesho/rv-iop/internal/api/handler"

type GetRvSubstituteCtFeedRequest struct {
	handler.GetRecentlyViewedFeedRequest
	//todo : check if this needs to be changed to RvSubstituteFeedRequest -> will change if it is not being used in other rv requests
}

type GetRvSubstituteCtFeedResponse struct {
	handler.GetRecentlyViewedFeedResponse
}
