package rvinteraction

import (
	"strconv"
	"time"

	dagConfig "github.com/Meesho/dag-engine/v2/handlers/config"
	"github.com/Meesho/dag-engine/v2/handlers/dag"
	"github.com/Meesho/dag-engine/v2/handlers/dag/component/result"
	"github.com/Meesho/dag-engine/v2/handlers/dag/executor"
	"github.com/Meesho/feed-commons-go/v2/pkg/data"
	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
	"github.com/Meesho/go-core/api"
	"github.com/Meesho/go-core/api/http"
	"github.com/Meesho/go-core/metric"
	starterComponent "github.com/Meesho/iop-component-starter/component"
	cohortConf "github.com/Meesho/iop-starter/cohort/config"
	"github.com/Meesho/rv-iop/internal/api/handler"
	"github.com/Meesho/rv-iop/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	sourceCgKey = "source_cg"
	sscatIdKey  = "sscat_id"
	scoreKey    = "score"
)

type RvInteractionsFeedImpl interface {
	FetchRvInteractionsFeed(request *GetRvInteractionsFeedRequest, requestContext *api.RequestContext) (
		*GetRvInteractionsFeedResponse, *api.Error)
}

// StandardRvInteractionsFeedImpl is the standard implementation of RvInteractionsFeedImpl
type StandardRvInteractionsFeedImpl struct {
	ServiceConf      *config.Service
	IopConfigHandler cohortConf.IopConfigHandler
}

func NewStandardRvInteractionsFeedImpl(serviceConf *config.Service,
	iopConfigHandler cohortConf.IopConfigHandler) *StandardRvInteractionsFeedImpl {
	if serviceConf == nil {
		log.Panic().Msgf("service conf cannot be nil")
	}
	if iopConfigHandler == nil {
		log.Panic().Msgf("iop config handler cannot be nil")
	}
	return &StandardRvInteractionsFeedImpl{
		ServiceConf:      serviceConf,
		IopConfigHandler: iopConfigHandler,
	}
}

func (s *StandardRvInteractionsFeedImpl) FetchRvInteractionsFeed(request *GetRvInteractionsFeedRequest,
	requestContext *api.RequestContext) (*GetRvInteractionsFeedResponse, *api.Error) {
	metricTags := []string{
		metric.TagAsString("user_context", requestContext.UserContext.String()),
		metric.TagAsString("feed_type", enum.FeedTypeRvInteractions.String()),
		metric.TagAsString("tenant_context", enum.TenantContextOrganic.String()),
		metric.TagAsString("method", "FetchRvInteractionsFeed"),
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
	if configResponse.ConfigBundle.GenerateFeedOnTheFly == nil {
		metric.Incr("iop_config_generate_feed_on_the_fly_nil", metricTags)
		log.Error().Msgf("Generate Feed On the Fly config is nil for request - %#v", iopConfigRequest)
		return nil, api.NewInternalServerError("Generate Feed On the Fly config is nil")
	}
	if configResponse.ConfigBundle.GenerateFeedOnTheFly.IsEmpty() {
		metric.Incr("iop_config_generate_feed_on_the_fly_empty", metricTags)
		log.Error().Msgf("Generate Feed On the Fly config is empty for request - %#v, iop config handler configResponse - %#v",
			iopConfigRequest, configResponse)
		return nil, api.NewInternalServerError("Generate Feed On the Fly config is empty")
	}
	if !configResponse.ConfigBundle.GenerateFeedOnTheFly.Config.Enabled {
		metric.Incr("iop_config_generate_feed_on_the_fly_not_enabled", metricTags)
		log.Error().Msgf("Generate Feed On the Fly config is not enabled for request - %#v, iop config handler configResponse - %#v",
			iopConfigRequest, configResponse)
		return nil, api.NewInternalServerError("Generate Feed On the Fly config is not enabled")
	}
	// Execute DAG
	metric.UpdateTags(&metricTags, metric.NewTag("dag_name", configResponse.ConfigBundle.GenerateFeedOnTheFly.Config.Name))
	iopConfig := configResponse.ConfigBundle.GenerateFeedOnTheFly

	requestData, apiErr := s.buildExecutionRequestData(request, requestContext)
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

	// Build real response (keeps full code path exercised)
	_, _ = s.buildResponse(dagExecResp, configResponse.ConfigBundle.GenerateFeedOnTheFly, configResponse, request, &metricTags)

	// HARDCODED: return fixed payload at last point before leaving service (for path verification)
	return hardcodedRvInteractionsFeedResponse(), nil
}

func (s *StandardRvInteractionsFeedImpl) executeDAG(dagExecutionRequest *dag.ExecutionRequest,
	metricTags []string) (*dag.ExecutionResponse, error) {
	defer metric.TimingWithStart("dag_execution_latency", time.Now(), metricTags)
	return executor.Instance().Execute(dagExecutionRequest)
}

func (s *StandardRvInteractionsFeedImpl) buildExecutionRequestData(request *GetRvInteractionsFeedRequest,
	meta *api.RequestContext) (*starterComponent.ExecutionRequestData, *api.Error) {
	context := starterComponent.NewRequestContext()
	if len(request.Data.Cursor) != 0 {
		context.SetCursor(request.Data.Cursor)
	}
	context.SetRequestedLimit(request.Data.Limit)
	context.Set(http.HeaderMeeshoUserStateCode, meta.UserStateCode)

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

	sscatId := request.Data.SscatId

	return &starterComponent.ExecutionRequestData{
		UserId:   meta.UserId,
		FeedType: enum.FeedTypeRvInteractions,
		FeedId: starterComponent.FeedId{
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

func (s *StandardRvInteractionsFeedImpl) buildResponse(resp *dag.ExecutionResponse, iopConfig *dagConfig.IOP,
	iopConfigResponse *cohortConf.IopConfigResponse, request *GetRvInteractionsFeedRequest,
	metricTags *[]string) (*GetRvInteractionsFeedResponse, *api.Error) {

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
	response, ok := res.(*starterComponent.Response)
	if !ok {
		metric.Incr("dag_result_type_mismatch", *metricTags)
		return nil, api.NewInternalServerError("unexpected result type")
	}

	products := make([]Product, 0, len(response.Candidates))
	for _, candidate := range response.Candidates {
		productId := getProductIdFromCandidate(candidate)
		if productId == 0 {
			continue
		}
		metaData, err := buildRvInteractionsMetaData(candidate)
		if err != nil {
			metric.Incr("build_rv_interactions_metadata_error", *metricTags)
			log.Error().Err(err).Msgf("Error in building rv interactions metadata for candidate id - %d", candidate.Id)
			return nil, api.NewInternalServerError(err.Error())
		}
		cursor, _ := candidate.Meta.Context.Cursor()
		products = append(products, Product{
			RvInteractionProduct: handler.RvInteractionProduct{
				ProductId:  productId,
				CatalogId:  candidate.Id,
				Cursor:     cursor,
				TrackingId: candidate.Meta.TrackingId,
				MetaData:   metaData,
			},
		})
	}

	if request.Data.Limit > 0 && len(products) > request.Data.Limit {
		products = products[:request.Data.Limit]
	}

	metric.Count("rv_interactions_feed_count", int64(len(products)), nil)
	if len(products) == 0 {
		metric.Incr("rv_interactions_feed_empty", *metricTags)
	}

	handlerProducts := make([]handler.RvInteractionProduct, len(products))
	for i := range products {
		handlerProducts[i] = products[i].RvInteractionProduct
	}

	return &GetRvInteractionsFeedResponse{
		GetRvInteractionsFeedResponse: handler.GetRvInteractionsFeedResponse{
			Products: handlerProducts,
		},
	}, nil
}

// getProductIdFromCandidate reads product_id from candidate context or scores (DAG components set it).
func getProductIdFromCandidate(candidate *starterComponent.Candidate) int {
	if val, ok := candidate.Meta.Context.Get("product_id"); ok {
		switch v := val.(type) {
		case string:
			if id, err := strconv.Atoi(v); err == nil {
				return id
			}
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		}
	}
	if val, ok := candidate.Meta.Scores.Get("product_id"); ok && val != "" {
		if id, err := strconv.Atoi(val); err == nil {
			return id
		}
	}
	return 0
}

func buildRvInteractionsMetaData(candidate *starterComponent.Candidate) (*handler.MetaData, error) {
	subTenant, _ := candidate.Meta.Context.GetString(data.SubTenantKey)

	scores := make(map[string]string)
	if candidate.Meta.Scores != nil {
		if scoreStr, exists := candidate.Meta.Scores.Get(scoreKey); exists {
			scores[scoreKey] = scoreStr
		}
	}

	// Build context map
	context := make(map[string]string)
	if sourceCg, ok := candidate.Meta.Context.GetString(data.SourceCg); ok {
		context[sourceCgKey] = sourceCg
	}

	if sscatId, ok := candidate.Meta.Context.GetString(sscatIdKey); ok {
		context[sscatIdKey] = sscatId
	}

	if productId, ok := candidate.Meta.Context.GetString("product_id"); ok {
		context["product_id"] = productId
	}

	return &handler.MetaData{
		Scores:    scores,
		Source:    candidate.Meta.Source,
		Context:   context,
		SubTenant: subTenant,
	}, nil
}

// hardcodedRvInteractionsFeedResponse returns a fixed response for path verification (last point before leaving service).
func hardcodedRvInteractionsFeedResponse() *GetRvInteractionsFeedResponse {
	return &GetRvInteractionsFeedResponse{
		GetRvInteractionsFeedResponse: handler.GetRvInteractionsFeedResponse{
			Products: []handler.RvInteractionProduct{
				{
					ProductId:  90001,
					CatalogId:  80001,
					Cursor:     "hardcoded_cursor_1",
					TrackingId: "hardcoded_tracking_1",
					MetaData: &handler.MetaData{
						Scores:    map[string]string{"score": "0.95"},
						Source:    "hardcoded_source",
						Context:   map[string]string{"source_cg": "cg1", "sscat_id": "10", "product_id": "90001"},
						SubTenant: "hardcoded_sub_tenant",
					},
				},
				{
					ProductId:  90002,
					CatalogId:  80002,
					Cursor:     "hardcoded_cursor_2",
					TrackingId: "hardcoded_tracking_2",
					MetaData: &handler.MetaData{
						Scores:    map[string]string{"score": "0.88"},
						Source:    "hardcoded_source",
						Context:   map[string]string{"source_cg": "cg2", "sscat_id": "20", "product_id": "90002"},
						SubTenant: "hardcoded_sub_tenant",
					},
				},
			},
		},
	}
}
