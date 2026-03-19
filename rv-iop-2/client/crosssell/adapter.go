package crosssell

import (
	"encoding/json"

	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
	grpc "github.com/Meesho/rv-iop/client/grpc/rv-iop"
	"github.com/rs/zerolog/log"
)

type Adapter struct{}

func (a *Adapter) mapToCrossSellWidgetRequestProto(request *CrossSellRequestData) (*grpc.GetCrossSellWidgetRequest, error) {
	return &grpc.GetCrossSellWidgetRequest{
		Request: &grpc.GetCrossSellWidgetRequestData{
			Data: &grpc.CrossSellRequestData{
				ParentEntityIds:  request.ParentEntityIds,
				ParentEntityType: request.ParentEntityType.String(),
				FeedContext:      request.FeedContext.String(),
				Cursor:           request.Cursor,
				Limit:            int32(request.Limit),
				Meta:             request.Meta,
				SubSubCategoryId: int32(request.SubSubCategoryId),
			},
		},
	}, nil
}

func (a *Adapter) mapCrossSellWidgetResponseFromProto(response *grpc.GetCrossSellWidgetResponse) *CrossSellResponseData {
	if response == nil || response.Response == nil || response.Response.Data == nil {
		return &CrossSellResponseData{
			SimilarEntities: []SimilarCandidatesResponse{},
		}
	}

	responseData := response.Response.Data
	similarEntities := make([]SimilarCandidatesResponse, 0, len(responseData.SimilarCandidates))

	for _, candidate := range responseData.SimilarCandidates {
		var meta interface{}
		if candidate.Meta != "" {
			err := json.Unmarshal([]byte(candidate.Meta), &meta)
			if err != nil {
				log.Warn().Msgf("error parsing meta data for candidate %d", candidate.Id)
			}
		}

		metaData := MetaData{}
		if candidate.Metadata != nil {
			metaData = MetaData{
				Scores:    candidate.Metadata.Scores,
				Source:    candidate.Metadata.Source,
				Context:   candidate.Metadata.Context,
				SubTenant: candidate.Metadata.SubTenant,
			}
		}

		similarEntities = append(similarEntities, SimilarCandidatesResponse{
			Id:         int(candidate.Id),
			Cursor:     candidate.Cursor,
			Meta:       meta,
			TrackingId: candidate.TrackingId,
			MetaData:   metaData,
		})
	}

	// Parse tenant context from string
	tenantContext, _ := enum.ParseTenantContext(responseData.TenantContext)

	return &CrossSellResponseData{
		TenantContext:   tenantContext,
		HasNextEntity:   responseData.HasNextEntity,
		SimilarEntities: similarEntities,
		Slots:           responseData.Slots,
	}
}
