package grpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
	"github.com/Meesho/go-core/api"
	iopcomponent "github.com/Meesho/iop-starter/component"
	rviop "github.com/Meesho/rv-iop/client/grpc/rv-iop"
	"github.com/Meesho/rv-iop/internal/api/handler"
	"github.com/Meesho/rv-iop/internal/api/handler/ad"
	"github.com/Meesho/rv-iop/internal/api/handler/crosssell"
	"github.com/Meesho/rv-iop/internal/api/handler/exploit"
	"github.com/Meesho/rv-iop/internal/api/handler/explore"
	rvinteractionhandler "github.com/Meesho/rv-iop/internal/api/handler/rv-interaction"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// FetchRvSubstituteExploitFeed implements the RvIopServiceServer interface for exploit feed
func (s *RvIopGrpcServer) FetchRvSubstituteExploitFeed(
	ctx context.Context, req *rviop.GetRvSubstituteOrganicFeedRequest,
) (*rviop.GetRvSubstituteOrganicFeedResponse, error) {
	requestContext, err := api.GetRequestContextForGRPC(ctx)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request context: %s", err))
	}
	requestData, err := convertProtoToRequestData(req.Request.Data)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request: %s", err))
	}
	request := &exploit.GetRvSubstituteOrganicFeedRequest{
		GetRecentlyViewedFeedRequest: handler.GetRecentlyViewedFeedRequest{
			Data: requestData,
		},
	}
	response, apiError := s.ExploitHandler.FetchRvSubstitute(request, requestContext)
	if apiError != nil {
		log.Error().Err(apiError).Msg("Error in exploit handler")
		return nil, api.ConvertHttpErrorToGrpc(apiError)
	}
	similarEntities := convertToProtoResponseData(&response.GetRecentlyViewedFeedResponse.Data)
	grpcResponse := &rviop.GetRvSubstituteOrganicFeedResponse{
		Response: &rviop.GetRecentlyViewedFeedResponse{
			Data:  similarEntities,
			Error: nil,
		},
	}
	return grpcResponse, nil
}

// FetchRvSubstituteExploreFeed implements the RvIopServiceServer interface for explore feed
func (s *RvIopGrpcServer) FetchRvSubstituteExploreFeed(
	ctx context.Context, req *rviop.GetRvSubstituteCtFeedRequest,
) (*rviop.GetRvSubstituteCtFeedResponse, error) {
	requestContext, err := api.GetRequestContextForGRPC(ctx)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request context: %s", err))
	}
	requestData, err := convertProtoToRequestData(req.Request.Data)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request: %s", err))
	}
	request := &explore.GetRvSubstituteCtFeedRequest{
		GetRecentlyViewedFeedRequest: handler.GetRecentlyViewedFeedRequest{
			Data: requestData,
		},
	}
	response, apiError := s.ExploreHandler.FetchRvSubstituteCt(request, requestContext)
	if apiError != nil {
		log.Error().Err(apiError).Msg("Error in ct handler")
		return nil, api.ConvertHttpErrorToGrpc(apiError)
	}
	similarEntities := convertToProtoResponseData(&response.GetRecentlyViewedFeedResponse.Data)
	grpcResponse := &rviop.GetRvSubstituteCtFeedResponse{
		Response: &rviop.GetRecentlyViewedFeedResponse{
			Data:  similarEntities,
			Error: nil,
		},
	}
	return grpcResponse, nil
}

// FetchRvSubstituteAdFeed implements the RvIopServiceServer interface for ads feed
func (s *RvIopGrpcServer) FetchRvSubstituteAdFeed(
	ctx context.Context, req *rviop.GetRvSubstituteAdFeedRequest,
) (*rviop.GetRvSubstituteAdFeedResponse, error) {
	requestContext, err := api.GetRequestContextForGRPC(ctx)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request context: %s", err))
	}
	requestData, err := convertProtoToRequestData(req.Request.Data)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request: %s", err))
	}
	request := &ad.GetRvSubstituteAdFeedRequest{
		GetRecentlyViewedFeedRequest: handler.GetRecentlyViewedFeedRequest{
			Data: requestData,
		},
	}
	response, apiError := s.AdHandler.FetchRvSubstituteAd(request, requestContext)
	if apiError != nil {
		log.Error().Err(apiError).Msg("Error in ad handler")
		return nil, api.ConvertHttpErrorToGrpc(apiError)
	}
	similarEntities := convertToProtoResponseData(&response.GetRecentlyViewedFeedResponse.Data)
	grpcResponse := &rviop.GetRvSubstituteAdFeedResponse{
		Response: &rviop.GetRecentlyViewedFeedResponse{
			Data:  similarEntities,
			Error: nil,
		},
	}
	return grpcResponse, nil
}

// FetchCrossSellWidget implements the RvIopServiceServer interface for cross-sell widget
func (s *RvIopGrpcServer) FetchCrossSellWidget(
	ctx context.Context, req *rviop.GetCrossSellWidgetRequest) (*rviop.GetCrossSellWidgetResponse, error) {
	requestContext, err := api.GetRequestContextForGRPC(ctx)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request context: %s", err))
	}

	requestData, err := convertProtoToCrossSellWidgetRequestData(req.Request.Data)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request: %s", err))
	}

	request := &crosssell.GetCrossSellWidgetRequest{
		GetCrossSellWidgetRequest: handler.GetCrossSellWidgetRequest{
			Data: requestData,
		},
	}

	response, apiError := s.CrossSellWidgetHandler.FetchCrossSellWidget(request, requestContext)
	if apiError != nil {
		log.Error().Err(apiError).Msg("Error in cross-sell widget handler")
		return nil, api.ConvertHttpErrorToGrpc(apiError)
	}
	crossSellEntities := convertToProtoResponseData(&response.GetCrossSellWidgetResponse.Data)

	grpcResponse := &rviop.GetCrossSellWidgetResponse{
		Response: &rviop.GetCrossSellWidgetResponseData{
			Data:  crossSellEntities,
			Error: nil,
		},
	}

	return grpcResponse, nil
}

// FetchCrossSellFeed implements the RvIopServiceServer interface for cross-sell feed
func (s *RvIopGrpcServer) FetchCrossSellFeed(
	ctx context.Context, req *rviop.GetCrossSellFeedRequest) (*rviop.GetCrossSellFeedResponse, error) {
	requestContext, err := api.GetRequestContextForGRPC(ctx)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request context: %s", err))
	}

	requestData, err := convertProtoToCrossSellFeedRequestData(req.Request.Data)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request: %s", err))
	}

	request := &crosssell.GetCrossSellFeedRequest{
		GetCrossSellFeedRequest: handler.GetCrossSellFeedRequest{
			Data: requestData,
		},
	}

	response, apiError := s.CrossSellFeedHandler.FetchCrossSellFeed(request, requestContext)
	if apiError != nil {
		log.Error().Err(apiError).Msg("Error in cross-sell feed handler")
		return nil, api.ConvertHttpErrorToGrpc(apiError)
	}
	crossSellEntities := convertToProtoResponseData(&response.GetCrossSellFeedResponse.Data)

	grpcResponse := &rviop.GetCrossSellFeedResponse{
		Response: &rviop.GetCrossSellFeedResponseData{
			Data:  crossSellEntities,
			Error: nil,
		},
	}

	return grpcResponse, nil
}

func (s *RvIopGrpcServer) GetComponents(ctx context.Context, req *emptypb.Empty) (*rviop.GetComponentsResponse, error) {
	components := iopcomponent.AllComponents()
	protoComponents := make([]*rviop.ComponentInfo, 0, len(components))
	for _, comp := range components {
		protoComponents = append(protoComponents, &rviop.ComponentInfo{
			Name:   comp.Name,
			Source: comp.Source,
		})
	}

	return &rviop.GetComponentsResponse{
		Components: protoComponents,
	}, nil
}

func convertProtoToRequestData(protoData *rviop.RequestData) (handler.RequestData, error) {
	feedContext, err := enum.ParseFeedContext(protoData.FeedContext)
	if err != nil {
		log.Err(err).Msg("Invalid feed context")
		return handler.RequestData{}, fmt.Errorf("invalid feed context")
	}
	return handler.RequestData{
		ParentEntityId:   int(protoData.ParentEntityId),
		ParentEntityType: enum.EntityTypeCatalog,
		FeedContext:      feedContext,
		Cursor:           protoData.Cursor,
		Limit:            int(protoData.Limit),
		Meta:             protoData.Meta,
		SubSubCategoryId: int(protoData.SubSubCategoryId),
	}, nil
}

// convertProtoToCrossSellFeedRequestData converts proto data to CrossSellRequestData for feed
func convertProtoToCrossSellFeedRequestData(protoData *rviop.CrossSellRequestData) (handler.CrossSellRequestData, error) {
	feedContext, err := enum.ParseFeedContext(protoData.FeedContext)
	if err != nil {
		log.Err(err).Msg("Invalid feed context")
		return handler.CrossSellRequestData{}, fmt.Errorf("invalid feed context")
	}

	if protoData.SubSubCategoryId <= 0 {
		return handler.CrossSellRequestData{}, fmt.Errorf("invalid sscat id")
	}

	parentEntityIds := make([]int, len(protoData.ParentEntityIds))
	for i, id := range protoData.ParentEntityIds {
		parentEntityIds[i] = int(id)
	}

	return handler.CrossSellRequestData{
		ParentEntityIds:  parentEntityIds,
		ParentEntityType: enum.EntityTypeCatalog,
		FeedContext:      feedContext,
		Cursor:           protoData.Cursor,
		Limit:            int(protoData.Limit),
		Meta:             protoData.Meta,
		SubSubCategoryId: int(protoData.SubSubCategoryId),
	}, nil
}

// convertProtoToCrossSellWidgetRequestData converts proto data to CrossSellRequestData for widget
func convertProtoToCrossSellWidgetRequestData(protoData *rviop.CrossSellRequestData) (handler.CrossSellRequestData, error) {
	feedContext, err := enum.ParseFeedContext(protoData.FeedContext)
	if err != nil {
		log.Err(err).Msg("Invalid feed context")
		return handler.CrossSellRequestData{}, fmt.Errorf("invalid feed context")
	}

	parentEntityIds := make([]int, len(protoData.ParentEntityIds))
	for i, id := range protoData.ParentEntityIds {
		parentEntityIds[i] = int(id)
	}

	return handler.CrossSellRequestData{
		ParentEntityIds:  parentEntityIds,
		ParentEntityType: enum.EntityTypeCatalog,
		FeedContext:      feedContext,
		Cursor:           protoData.Cursor,
		Limit:            int(protoData.Limit),
		Meta:             protoData.Meta,
		SubSubCategoryId: int(protoData.SubSubCategoryId),
	}, nil
}

func convertToProtoResponseData(internalData *handler.ResponseData) *rviop.ResponseData {
	similarCandidates := make([]*rviop.SimilarCandidatesResponse, 0, len(internalData.SimilarEntities))
	for _, entity := range internalData.SimilarEntities {
		metaStr := ""
		if entity.Meta != nil {
			metaStr, _ = entity.Meta.(string)
		}
		similarCandidates = append(similarCandidates, &rviop.SimilarCandidatesResponse{
			Id:         int32(entity.Id),
			Cursor:     entity.Cursor,
			TrackingId: entity.TrackingId,
			Meta:       metaStr,
			Metadata:   convertToProtoMetadata(&entity.MetaData),
		})
	}
	return &rviop.ResponseData{
		TenantContext:     internalData.TenantContext.String(),
		HasNextEntity:     internalData.HasNextEntity,
		SimilarCandidates: similarCandidates,
		Slots:             internalData.Slots,
	}
}
func convertToProtoMetadata(internalMeta *handler.MetaData) *rviop.MetaData {
	if internalMeta == nil {
		return nil
	}
	return &rviop.MetaData{
		Scores:    internalMeta.Scores,
		Source:    internalMeta.Source,
		Context:   internalMeta.Context,
		SubTenant: internalMeta.SubTenant,
		Sse:       internalMeta.Sse,
	}
}

// FetchRvInteractionsFeed implements the RvIopServiceServer interface for rv interactions feed
func (s *RvIopGrpcServer) FetchRvInteractionsFeed(
	ctx context.Context, req *rviop.GetRvInteractionsFeedRequest,
) (*rviop.GetRvInteractionsFeedResponse, error) {
	requestContext, err := api.GetRequestContextForGRPC(ctx)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request context: %s", err))
	}

	if req == nil || req.Request == nil || req.Request.Data == nil {
		return nil, api.NewGrpcInvalidArgumentError("Request or request data is nil")
	}

	requestData, err := convertProtoToRvInteractionsRequestData(req.Request.Data)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request: %s", err))
	}

	request := &rvinteractionhandler.GetRvInteractionsFeedRequest{
		GetRvInteractionsFeedRequest: handler.GetRvInteractionsFeedRequest{
			Data: requestData.RvInteractionsRequestData,
		},
	}

	response, apiError := s.RvInteractionsFeedHandler.FetchRvInteractionsFeed(request, requestContext)
	if apiError != nil {
		log.Error().Err(apiError).Msg("Error in rv interactions feed handler")
		return nil, api.ConvertHttpErrorToGrpc(apiError)
	}

	// Convert Products to SimilarCandidatesResponse format
	similarCandidates := make([]*rviop.SimilarCandidatesResponse, 0, len(response.Products))
	for _, product := range response.Products {
		similarCandidates = append(similarCandidates, &rviop.SimilarCandidatesResponse{
			Id:         int32(product.CatalogId),
			Cursor:     product.Cursor,
			TrackingId: product.TrackingId,
			Meta:       "",
			Metadata:   convertToProtoMetadata(product.MetaData),
		})
	}

	grpcResponse := &rviop.GetRvInteractionsFeedResponse{
		Response: &rviop.GetRecentlyViewedInteractionsFeedResponse{
			Data: &rviop.ResponseData{
				TenantContext:     enum.TenantContextOrganic.String(),
				HasNextEntity:     false,
				SimilarCandidates: similarCandidates,
			},
			Error: nil,
		},
	}

	return grpcResponse, nil
}

// FetchRvInteractionsWidget implements the RvIopServiceServer interface for rv interaction widget
func (s *RvIopGrpcServer) FetchRvInteractionsWidget(
	ctx context.Context, req *rviop.GetRvInteractionsWidgetRequest,
) (*rviop.GetRvInteractionsWidgetResponse, error) {
	requestContext, err := api.GetRequestContextForGRPC(ctx)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request context: %s", err))
	}

	if req == nil || req.Request == nil || req.Request.Data == nil {
		return nil, api.NewGrpcInvalidArgumentError("Request or request data is nil")
	}

	// Convert proto request to internal request
	data, err := convertProtoToRvInteractionsRequestData(req.Request.Data)
	if err != nil {
		return nil, api.NewGrpcInvalidArgumentError(fmt.Sprintf("Invalid request: %s", err))
	}
	internalReq := &rvinteractionhandler.GetRvInteractionsWidgetRequest{
		GetRvInteractionsWidgetRequest: handler.GetRvInteractionsWidgetRequest{
			Data: data.RvInteractionsRequestData,
		},
	}

	// Delegate to handler
	widgetResp, apiErr := s.RvInteractionsWidgetHandler.FetchRvInteractionsWidget(internalReq, requestContext)
	if apiErr != nil {
		return nil, api.ConvertHttpErrorToGrpc(apiErr)
	}

	// Build productId -> product map for lookup (includes catalogId, cursor, trackingId)
	productIdToProduct := make(map[int]*handler.RvInteractionProduct)
	for i := range widgetResp.Products {
		product := &widgetResp.Products[i]
		productIdToProduct[product.ProductId] = product
	}

	const (
		metadataKeyProductId = "product_id"
		metadataKeySscatId   = "sscat_id"
		metadataKeySscatName = "sscat_name"
	)

	similarCandidates := make([]*rviop.SimilarCandidatesResponse, 0)
	for _, category := range widgetResp.Categories {
		for _, productId := range category.ProductIds {
			product := productIdToProduct[productId]
			if product == nil {
				log.Error().Msgf("Product not found for product_id %d, skipping", productId)
				continue
			}
			similarCandidates = append(similarCandidates, &rviop.SimilarCandidatesResponse{
				Id:         int32(product.CatalogId),
				Cursor:     product.Cursor,
				TrackingId: product.TrackingId,
				Metadata: &rviop.MetaData{
					Context: map[string]string{
						metadataKeyProductId: strconv.Itoa(productId),
						metadataKeySscatId:   strconv.Itoa(category.SscatId),
						metadataKeySscatName: category.SscatName,
					},
				},
			})
		}
	}

	grpcResponse := &rviop.GetRvInteractionsWidgetResponse{
		Response: &rviop.GetRecentlyViewedInteractionsWidgetResponse{
			Data: &rviop.ResponseData{
				TenantContext:     enum.TenantContextOrganic.String(),
				HasNextEntity:     false,
				SimilarCandidates: similarCandidates,
			},
			Error: nil,
		},
	}
	return grpcResponse, nil
}

// convertProtoToRvInteractionsRequestData converts proto data to RvInteractionsRequestData
func convertProtoToRvInteractionsRequestData(protoData *rviop.RvInteractionsRequestData) (rvinteractionhandler.RvInteractionsRequestData, error) {
	return rvinteractionhandler.RvInteractionsRequestData{
		RvInteractionsRequestData: handler.RvInteractionsRequestData{
			UserId:      protoData.UserId,
			SscatId:     int(protoData.SscatId),
			Limit:       int(protoData.Limit),
			Cursor:      protoData.Cursor,
			Meta:        protoData.Meta,
			FeedContext: protoData.FeedContext,
		},
	}, nil
}
