package client

type PdpFeedRequest struct {
	CatalogId   int32
	Limit       int32
	Offset      int32
	FeedContext string
	Metadata    RequestMetadata
	Headers     map[string]string
}

type RequestMetadata struct {
	WidgetPosition      int32
	Source              string
	SessionId           string
	Theme               string
	WidgetGroupMetadata map[string]interface{}
}

type PdpFeedResponse struct {
	Catalogs []CatalogDTO
	Cursor   string
	Error    string
}

type CatalogDTO struct {
	Id        int32
	Source    string
	Tracker   TrackerDTO
	CatalogAd CatalogAdDTO
}

type TrackerDTO struct {
	TrackerId string
}

type CatalogAdDTO struct {
	Enabled    bool
	CampaignId int32
	Metadata   string
}
