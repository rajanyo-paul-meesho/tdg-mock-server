package rvinteraction

import "github.com/Meesho/rv-iop/internal/api/handler"

type GetRvInteractionsFeedRequest struct {
	handler.GetRvInteractionsFeedRequest
}

type Product struct {
	handler.RvInteractionProduct
}

type GetRvInteractionsWidgetRequest struct {
	handler.GetRvInteractionsWidgetRequest
}

type GetRvInteractionsWidgetResponse struct {
	handler.GetRvInteractionsWidgetResponse
}

type RecentlyViewedCategory struct {
	handler.RecentlyViewedCategory
}

type RvInteractionsRequestData struct {
	handler.RvInteractionsRequestData
}

type GetRvInteractionsFeedResponse struct {
	handler.GetRvInteractionsFeedResponse
}
