package rvinteraction

import (
	"strconv"
	"time"

	dagConfig "github.com/Meesho/dag-engine/v2/handlers/config"
	"github.com/Meesho/dag-engine/v2/handlers/dag"
	"github.com/Meesho/dag-engine/v2/handlers/dag/component/result"
	"github.com/Meesho/dag-engine/v2/handlers/dag/executor"
	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
	"github.com/Meesho/go-core/api"
	"github.com/Meesho/go-core/api/http"
	"github.com/Meesho/go-core/metric"
	component "github.com/Meesho/iop-component-starter/component"
	cohortConf "github.com/Meesho/iop-starter/cohort/config"
	"github.com/Meesho/rv-iop/internal/api/handler"
	"github.com/Meesho/rv-iop/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type RvInteractionsWidgetImpl interface {
	FetchRvInteractionsWidget(request *GetRvInteractionsWidgetRequest, requestContext *api.RequestContext) (
		*GetRvInteractionsWidgetResponse, *api.Error)
}

// StandardRvInteractionsWidgetImpl is the standard implementation of RvInteractionsWidgetImpl
type StandardRvInteractionsWidgetImpl struct {
	ServiceConf      *config.Service
	IopConfigHandler cohortConf.IopConfigHandler
}

func NewStandardRvInteractionsWidgetImpl(serviceConf *config.Service,
	iopConfigHandler cohortConf.IopConfigHandler) *StandardRvInteractionsWidgetImpl {
	if serviceConf == nil {
		log.Panic().Msgf("service conf cannot be nil")
	}
	if iopConfigHandler == nil {
		log.Panic().Msgf("iop config handler cannot be nil")
	}
	return &StandardRvInteractionsWidgetImpl{
		ServiceConf:      serviceConf,
		IopConfigHandler: iopConfigHandler,
	}
}

func (s *StandardRvInteractionsWidgetImpl) FetchRvInteractionsWidget(request *GetRvInteractionsWidgetRequest,
	requestContext *api.RequestContext) (*GetRvInteractionsWidgetResponse, *api.Error) {
	metricTags := []string{
		metric.TagAsString("user_context", requestContext.UserContext.String()),
		metric.TagAsString("feed_type", enum.FeedTypeRvInteractions.String()),
		metric.TagAsString("tenant_context", enum.TenantContextOrganic.String()),
		metric.TagAsString("method", "FetchRvInteractionsWidget"),
	}

	// Parse feed context or use default
	feedContext, err := enum.ParseFeedContext(request.Data.FeedContext)
	if err != nil {
		feedContext = enum.FeedContextDefault
	}

	metricTags = append(metricTags, metric.TagAsString("feed_context", feedContext.String()))

	iopConfigRequest := &cohortConf.IopConfigRequest{
		UserId:        requestContext.UserId,
		UserContext:   requestContext.UserContext,
		FeedType:      enum.FeedTypeRvInteractions,
		FeedContext:   feedContext,
		TenantContext: enum.TenantContextOrganic,
		ServiceTag:    s.ServiceConf.App.Name,
	}
	// Get IOP config from experiment handler
	configResponse, err := s.IopConfigHandler.GetConfig(iopConfigRequest)
	if err != nil {
		metric.Incr("iop_config_handler_error", metricTags)
		log.Error().Err(err).Msgf("Error in getting iop config for request - %#v", iopConfigRequest)
		return nil, api.NewInternalServerError(err.Error())
	}
	if configResponse.ConfigBundle.GenerateWidgetOnTheFly == nil {
		metric.Incr("iop_config_generate_widget_on_the_fly_nil", metricTags)
		log.Error().Msgf("Generate Widget On the Fly config is nil for request - %#v", iopConfigRequest)
		return nil, api.NewInternalServerError("Generate Widget On the Fly config is nil")
	}
	if configResponse.ConfigBundle.GenerateWidgetOnTheFly.IsEmpty() {
		metric.Incr("iop_config_generate_widget_on_the_fly_empty", metricTags)
		log.Error().Msgf("Generate Widget On the Fly config is empty for request - %#v, iop config handler configResponse - %#v",
			iopConfigRequest, configResponse)
		return nil, api.NewInternalServerError("Generate Widget On the Fly config is empty")
	}
	if !configResponse.ConfigBundle.GenerateWidgetOnTheFly.Config.Enabled {
		metric.Incr("iop_config_generate_widget_on_the_fly_not_enabled", metricTags)
		log.Error().Msgf("Generate Widget On the Fly config is not enabled for request - %#v, iop config handler configResponse - %#v",
			iopConfigRequest, configResponse)
		return nil, api.NewInternalServerError("Generate Widget On the Fly config is not enabled")
	}
	// Execute DAG
	metric.UpdateTags(&metricTags, metric.NewTag("dag_name", configResponse.ConfigBundle.GenerateWidgetOnTheFly.Config.Name))
	iopConfig := configResponse.ConfigBundle.GenerateWidgetOnTheFly

	requestData, apiErr := s.buildExecutionRequestData(request, requestContext, metricTags)
	if apiErr != nil {
		return nil, apiErr
	}

	dagExecResp, err := s.executeDAG(&dag.ExecutionRequest{
		IOPConfig: iopConfig,
		Data:      requestData,
		Meta: &dag.Meta{
			DagName:          iopConfig.Config.Name,
			VariantName:      configResponse.ConfigVariantName,
			ExpName:          configResponse.CohortName,
			IsLoggingEnabled: configResponse.IsLoggingEnabled,
		},
	}, metricTags)

	metric.Incr("dag_execution_total", metricTags)
	if err != nil {
		metric.Incr("dag_execution_error", metricTags)
		log.Error().Err(err).Msgf("Error in executing dag request - %#v", iopConfigRequest)
		return nil, api.NewInternalServerError(err.Error())
	}
	metric.Incr("dag_execution_success", metricTags)

	return s.buildResponse(dagExecResp, configResponse.ConfigBundle.GenerateWidgetOnTheFly, request, &metricTags)
}

func (s *StandardRvInteractionsWidgetImpl) executeDAG(dagExecutionRequest *dag.ExecutionRequest,
	metricTags []string) (*dag.ExecutionResponse, error) {
	defer metric.TimingWithStart("dag_execution_latency", time.Now(), metricTags)
	return executor.Instance().Execute(dagExecutionRequest)
}

func (s *StandardRvInteractionsWidgetImpl) buildExecutionRequestData(request *GetRvInteractionsWidgetRequest,
	meta *api.RequestContext, metricTags []string) (*component.ExecutionRequestData, *api.Error) {
	context := component.NewRequestContext()
	if len(request.Data.Cursor) != 0 {
		context.SetCursor(request.Data.Cursor)
	}
	context.SetRequestedLimit(request.Data.Limit)
	context.Set(http.HeaderMeeshoUserStateCode, meta.UserStateCode)

	// Set category limit if provided
	if request.Data.CategoryLimit != nil && *request.Data.CategoryLimit > 0 {
		context.Set("category_limit", strconv.Itoa(*request.Data.CategoryLimit))
	}

	// Set product per category limit if provided
	if request.Data.ProductPerCategoryLimit != nil && *request.Data.ProductPerCategoryLimit > 0 {
		context.Set("products_per_category_limit", strconv.Itoa(*request.Data.ProductPerCategoryLimit))
	}

	// Set additional context data from the request
	if request.Data.Meta != nil {
		for key, value := range request.Data.Meta {
			context.Set(key, value)
		}
	}

	// Parse feed context or use default
	feedContext, err := enum.ParseFeedContext(request.Data.FeedContext)
	if err != nil {
		feedContext = enum.FeedContextDefault
	}

	// We set CatalogId to nil and SSCatId to 0 to fetch all interactions
	sscatId := 0

	return &component.ExecutionRequestData{
		UserId:   meta.UserId,
		FeedType: enum.FeedTypeRvInteractions,
		FeedId: component.FeedId{
			CatalogId: nil,
			SSCatId:   &sscatId,
		},
		FeedMetaData:  request.Data.Meta,
		FeedContext:   feedContext,
		UserContext:   meta.UserContext,
		IopId:         uuid.New().String(),
		TenantContext: enum.TenantContextOrganic,
		Context:       context,
		EntityType:    enum.EntityTypeCatalog,
	}, nil
}

func (s *StandardRvInteractionsWidgetImpl) buildResponse(resp *dag.ExecutionResponse, iopConfig *dagConfig.IOP,
	request *GetRvInteractionsWidgetRequest,
	metricTags *[]string) (*GetRvInteractionsWidgetResponse, *api.Error) {

	// fetch result from the result component mentioned in the iop config
	resFuture, ok := resp.Results[iopConfig.Config.ResultComponent].(*result.Future)
	if !ok {
		metric.Incr("dag_result_component_missing", *metricTags)
		return nil, api.NewInternalServerError("result component missing or invalid")
	}
	res, err := resFuture.Get()
	if err != nil {
		metric.Incr("dag_result_fetch_error", *metricTags)
		log.Error().Err(err).Msgf("Error in getting result from result component - %s", iopConfig.Config.ResultComponent)
		return nil, api.NewInternalServerError(err.Error())
	}
	metric.Incr("dag_result_fetch_success", *metricTags)
	response, ok := res.(*component.Response)
	if !ok {
		metric.Incr("dag_result_type_mismatch", *metricTags)
		return nil, api.NewInternalServerError("unexpected result type")
	}

	// Build products list matching feed logic exactly
	products := make([]Product, 0, len(response.Candidates))
	for _, candidate := range response.Candidates {
		// Extract product_id from context or scores
		productId := 0
		if val, ok := candidate.Meta.Context.Get("product_id"); ok {
			switch v := val.(type) {
			case string:
				if id, err := strconv.Atoi(v); err == nil {
					productId = id
				}
			case int:
				productId = v
			case int32:
				productId = int(v)
			case int64:
				productId = int(v)
			}
		}

		// Fallback: check Scores if not in Context (shouldn't happen after rv_sscat_populator, but for safety)
		if productId == 0 {
			if val, ok := candidate.Meta.Scores.Get("product_id"); ok && val != "" {
				if id, err := strconv.Atoi(val); err == nil {
					productId = id
				}
			}
		}

		// Build meta data from candidate context and scores
		metaData, err := buildRvInteractionsMetaData(candidate)
		if err != nil {
			metric.Incr("build_rv_interactions_metadata_error", *metricTags)
			log.Error().Err(err).Msgf("Error in building rv interactions metadata for candidate id - %d", candidate.Id)
			return nil, api.NewInternalServerError(err.Error())
		}

		// Extract cursor and trackingId from candidate (required fields)
		cursor, ok := candidate.Meta.Context.Cursor()
		if !ok {
			log.Error().Msgf("cursor is not set for candidate %d, using empty string", candidate.Id)
			cursor = ""
		}
		trackingId := candidate.Meta.TrackingId

		products = append(products, Product{
			RvInteractionProduct: handler.RvInteractionProduct{
				ProductId:  productId,
				CatalogId:  candidate.Id,
				Cursor:     cursor,
				TrackingId: trackingId,
				MetaData:   metaData,
			},
		})
		log.Debug().Msgf("Added product_id %d for candidate %d (catalog_id: %d)",
			productId, candidate.Id, candidate.Id)
	}

	// Apply limit if needed
	beforeLimitCount := len(products)
	if request.Data.Limit > 0 && len(products) > request.Data.Limit {
		products = products[:request.Data.Limit]
	}

	// Log detailed statistics for debugging
	log.Info().
		Int("products_before_limit", beforeLimitCount).
		Int("products_after_limit", len(products)).
		Int("request_limit", request.Data.Limit).
		Msg("RV interactions widget filtering statistics")

	tags := append(*metricTags, metric.TagAsString("widget_length", "rv_interactions"))
	metric.Histogram(metric.MethodLatency, float64(len(products)), tags)

	metric.Count("rv_interaction_widget_count", int64(len(products)), nil)
	if len(products) == 0 {
		metric.Incr("rv_interaction_widget_empty", *metricTags)
	}

	// Group products by sscat_id to build categories
	categories := s.buildCategoriesFromProducts(products, request)

	// Convert []Product to []handler.RvInteractionProduct for the embedded struct
	handlerProducts := make([]handler.RvInteractionProduct, len(products))
	for i := range products {
		handlerProducts[i] = products[i].RvInteractionProduct
	}

	// Convert []RecentlyViewedCategory to []handler.RecentlyViewedCategory for the embedded struct
	handlerCategories := make([]handler.RecentlyViewedCategory, len(categories))
	for i := range categories {
		handlerCategories[i] = categories[i].RecentlyViewedCategory
	}

	return &GetRvInteractionsWidgetResponse{
		GetRvInteractionsWidgetResponse: handler.GetRvInteractionsWidgetResponse{
			Categories: handlerCategories,
			Products:   handlerProducts, // Include products for catalogId lookup in gRPC
		},
	}, nil
}

func (s *StandardRvInteractionsWidgetImpl) buildCategoriesFromProducts(
	products []Product,
	request *GetRvInteractionsWidgetRequest) []RecentlyViewedCategory {

	// Group products by sscat_id from metadata
	sscatMap := make(map[int]*RecentlyViewedCategory)
	orderedSscatIds := make([]int, 0)

	for _, product := range products {
		// Get sscat_id from metadata context
		sscatId := 0
		sscatName := ""
		if product.MetaData != nil && product.MetaData.Context != nil {
			if sscatIdStr, ok := product.MetaData.Context["sscat_id"]; ok {
				if id, err := strconv.Atoi(sscatIdStr); err == nil {
					sscatId = id
				}
			}
			if name, ok := product.MetaData.Context["sscat_name"]; ok {
				sscatName = name
			}
		}

		if sscatId == 0 {
			continue
		}

		// Initialize category if not exists
		if sscatMap[sscatId] == nil {
			sscatMap[sscatId] = &RecentlyViewedCategory{
				RecentlyViewedCategory: handler.RecentlyViewedCategory{
					SscatId:    sscatId,
					SscatName:  sscatName,
					ProductIds: make([]int, 0),
				},
			}
			orderedSscatIds = append(orderedSscatIds, sscatId)
		}

		// Add product_id to category
		sscatMap[sscatId].ProductIds = append(sscatMap[sscatId].ProductIds, product.ProductId)
	}

	// Convert map to slice using ordered keys
	categories := make([]RecentlyViewedCategory, 0, len(orderedSscatIds))
	for _, sscatId := range orderedSscatIds {
		category := sscatMap[sscatId]
		// Apply product per category limit if specified
		if request.Data.ProductPerCategoryLimit != nil && *request.Data.ProductPerCategoryLimit > 0 {
			limit := *request.Data.ProductPerCategoryLimit
			if len(category.ProductIds) > limit {
				category.ProductIds = category.ProductIds[:limit]
			}
		}
		categories = append(categories, *category)
	}

	// Apply category limit if specified
	if request.Data.CategoryLimit != nil && *request.Data.CategoryLimit > 0 {
		limit := *request.Data.CategoryLimit
		if len(categories) > limit {
			categories = categories[:limit]
		}
	}

	return categories
}
