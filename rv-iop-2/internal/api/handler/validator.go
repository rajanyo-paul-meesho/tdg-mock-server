package handler

import (
	"errors"

	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
)

func IsValidGetRecentlyViewedFeedRequest(request GetRecentlyViewedFeedRequest) (bool, error) {
	if request.Data.Limit <= 0 {
		return false, errors.New("limit should be positive")
	}
	// Make SSCat_Id check optional for wishlist feedContext
	if request.Data.FeedContext == enum.FeedContextWishlist {
		return true, nil
	}
	if request.Data.SubSubCategoryId <= 0 {
		return false, errors.New("invalid sscat id")
	}
	return true, nil
}

func IsValidGetCrossSellWidgetRequest(request GetCrossSellWidgetRequest) (bool, error) {
	if request.Data.Limit <= 0 {
		return false, errors.New("limit should be positive")
	}

	if len(request.Data.ParentEntityIds) == 0 {
		return false, errors.New("invalid parent entity id")
	}

	return true, nil
}

func IsValidGetCrossSellFeedRequest(request GetCrossSellFeedRequest) (bool, error) {
	if request.Data.Limit <= 0 {
		return false, errors.New("limit should be positive")
	}

	if len(request.Data.ParentEntityIds) == 0 {
		return false, errors.New("invalid parent entity id")
	}

	if request.Data.SubSubCategoryId <= 0 {
		return false, errors.New("invalid sscat id")
	}
	return true, nil
}
