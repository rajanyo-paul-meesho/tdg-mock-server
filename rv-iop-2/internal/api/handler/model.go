package handler

import (
	"github.com/Meesho/dag-engine/v2/pkg/config"
	data2 "github.com/Meesho/feed-commons-go/v2/pkg/data"
	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
	"github.com/Meesho/iop-component-starter/data"
)

type GetRecentlyViewedFeedRequest struct {
	Data RequestData `json:"data"`
}

type RequestData struct {
	ParentEntityId   int               `json:"parent_entity_id"`
	ParentEntityType enum.EntityType   `json:"parent_entity_type"`
	FeedContext      enum.FeedContext  `json:"feed_context"`
	Cursor           string            `json:"cursor"`
	Limit            int               `json:"limit"`
	Meta             map[string]string `json:"meta"`
	SubSubCategoryId int               `json:"sub_sub_category_id"`
}

type CrossSellRequestData struct {
	ParentEntityIds  []int             `json:"parent_entity_ids"`
	ParentEntityType enum.EntityType   `json:"parent_entity_type"`
	FeedContext      enum.FeedContext  `json:"feed_context"`
	Cursor           string            `json:"cursor"`
	Limit            int               `json:"limit"`
	Meta             map[string]string `json:"meta"`
	SubSubCategoryId int               `json:"sub_sub_category_id"`
}

type GetRecentlyViewedFeedResponse struct {
	Data  ResponseData `json:"data"`
	Error Error        `json:"error"`
}

type ResponseData struct {
	TenantContext   enum.TenantContext          `json:"tenant_context"`
	HasNextEntity   bool                        `json:"has_next_entity"`
	SimilarEntities []SimilarCandidatesResponse `json:"similar_entities"`
	Slots           []int32                     `json:"slots,omitempty"`
}

type SimilarCandidatesResponse struct {
	Id         int         `json:"id"`
	Cursor     string      `json:"cursor"`
	Meta       interface{} `json:"meta"`
	TrackingId string      `json:"tracking_id"`
	MetaData   MetaData    `json:"meta_data,omitempty"`
}

type MetaData struct {
	Scores    map[string]string `json:"scores"`
	Source    string            `json:"source"`
	Context   map[string]string `json:"context"`
	SubTenant string            `json:"sub_tenant"`
	Sse       data2.Sse         `json:"sse"`
}

type Error struct {
	Message string `json:"message"`
}

type CandidateResponseMetaData struct {
	Scores  data.Scores   `json:"scores"`
	Context config.AnyVal `json:"context"`
}

type GetCrossSellWidgetRequest struct {
	Data CrossSellRequestData `json:"data"`
}

type GetCrossSellWidgetResponse struct {
	Data  ResponseData `json:"data"`
	Error Error        `json:"error"`
}

type GetCrossSellFeedRequest struct {
	Data CrossSellRequestData `json:"data"`
}

type GetCrossSellFeedResponse struct {
	Data  ResponseData `json:"data"`
	Error Error        `json:"error"`
}

// ===================== RV Interactions (shared models) =====================

type RvInteractionsRequestData struct {
	UserId                  string            `json:"user_id"`
	SscatId                 int               `json:"sscat_id"`
	Limit                   int               `json:"limit"`
	Cursor                  string            `json:"cursor,omitempty"`
	Meta                    map[string]string `json:"meta,omitempty"`
	FeedContext             string            `json:"feed_context,omitempty"`
	CategoryLimit           *int              `json:"category_limit,omitempty"`
	ProductPerCategoryLimit *int              `json:"product_per_category_limit,omitempty"`
}

type GetRvInteractionsFeedRequest struct {
	Data RvInteractionsRequestData `json:"data"`
}

type RvInteractionProduct struct {
	ProductId  int       `json:"product_id"`
	CatalogId  int       `json:"catalog_id"`
	Cursor     string    `json:"cursor"`
	TrackingId string    `json:"tracking_id"`
	MetaData   *MetaData `json:"meta_data,omitempty"`
}

type GetRvInteractionsFeedResponse struct {
	Products []RvInteractionProduct `json:"products"`
}

type GetRvInteractionsWidgetRequest struct {
	Data RvInteractionsRequestData `json:"data"`
}

type RecentlyViewedCategory struct {
	SscatId    int    `json:"sscat_id"`
	SscatName  string `json:"sscat_name"`
	ProductIds []int  `json:"product_ids"`
}

type GetRvInteractionsWidgetResponse struct {
	Categories []RecentlyViewedCategory `json:"categories"`
	Products   []RvInteractionProduct   `json:"products,omitempty"`
}
