package crosssell

import (
	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
)

type CrossSellRequestData struct {
	ParentEntityIds  []int32           `json:"parent_entity_ids"`
	ParentEntityType enum.EntityType   `json:"parent_entity_type"`
	FeedContext      enum.FeedContext  `json:"feed_context"`
	Cursor           string            `json:"cursor"`
	Limit            int               `json:"limit"`
	Meta             map[string]string `json:"meta"`
	SubSubCategoryId int               `json:"sub_sub_category_id"`
	Headers          map[string]string `json:"headers"`
}

type CrossSellResponseData struct {
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
}
