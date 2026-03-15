package client

import (
	grpc "github.com/Meesho/feed-aggregator-go/client/grpc/pdp"
	"google.golang.org/protobuf/types/known/structpb"
)

type Adapter struct{}

func (a *Adapter) mapToPdpFeedRequestProto(request *PdpFeedRequest) (*grpc.RecommendationsRequest, error) {
	// Convert widget group metadata to structpb.Struct
	widgetGroupStruct, err := structpb.NewStruct(request.Metadata.WidgetGroupMetadata)
	if err != nil {
		return nil, err
	}

	// Determine feed context - for similar items, we use specific context
	feedContext := request.FeedContext
	if feedContext == "" {
		feedContext = request.Metadata.Theme
	}

	return &grpc.RecommendationsRequest{
		Cursor:      "",
		Offset:      request.Offset,
		Limit:       request.Limit,
		FeedContext: feedContext,
		CatalogId:   request.CatalogId,
		Metadata: &grpc.RequestMetadata{
			WidgetPosition:      request.Metadata.WidgetPosition,
			Source:              request.Metadata.Source,
			SessionId:           request.Metadata.SessionId,
			WidgetGroupMetadata: widgetGroupStruct,
			Theme:               request.Metadata.Theme,
		},
	}, nil
}

func (a *Adapter) mapFromPdpFeedResponseProto(response *grpc.RecommendationResponse) *PdpFeedResponse {
	if response == nil {
		return &PdpFeedResponse{
			Error: "empty response from server",
		}
	}

	catalogs := make([]CatalogDTO, len(response.Catalogs))
	for i, catalog := range response.Catalogs {
		catalogs[i] = CatalogDTO{
			Id:     catalog.Id,
			Source: catalog.Source,
			Tracker: TrackerDTO{
				TrackerId: catalog.Tracker.TrackerId,
			},
			CatalogAd: CatalogAdDTO{
				Enabled:    catalog.CatalogAd.Enabled,
				CampaignId: catalog.CatalogAd.CampaignId,
				Metadata:   catalog.CatalogAd.Metadata,
			},
		}
	}

	return &PdpFeedResponse{
		Catalogs: catalogs,
		Cursor:   response.Cursor,
		Error:    response.Error,
	}
}
