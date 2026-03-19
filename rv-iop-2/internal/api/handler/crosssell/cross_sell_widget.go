package crosssell

import (
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
	sourceCgKey        = "source_cg"
	sscatIdKey         = "sscat_id"
	groupIdKey         = "group_id"
	sscatNameKey       = "sscat_name"
	scoreKey           = "score"
	parentProductIdKey = "parent_product_id"
	parentCatalogIdKey = "parent_catalog_id"
	productIdKey       = "product_id"
)

type CrossSellWidgetImpl interface {
	FetchCrossSellWidget(request *GetCrossSellWidgetRequest, requestContext *api.RequestContext) (
		*GetCrossSellWidgetResponse, *api.Error)
}

// StandardCrossSellWidgetImpl is the standard implementation of CrossSellWidgetImpl
type StandardCrossSellWidgetImpl struct {
	ServiceConf      *config.Service
	IopConfigHandler cohortConf.IopConfigHandler
}

func NewStandardCrossSellWidgetImpl(serviceConf *config.Service,
	iopConfigHandler cohortConf.IopConfigHandler) *StandardCrossSellWidgetImpl {
	if serviceConf == nil {
		log.Panic().Msgf("service conf cannot be nil")
	}
	if iopConfigHandler == nil {
		log.Panic().Msgf("iop config handler cannot be nil")
	}
	return &StandardCrossSellWidgetImpl{
		ServiceConf:      serviceConf,
		IopConfigHandler: iopConfigHandler,
	}
}

func (s *StandardCrossSellWidgetImpl) FetchCrossSellWidget(request *GetCrossSellWidgetRequest,
	requestContext *api.RequestContext) (*GetCrossSellWidgetResponse, *api.Error) {
	metricTags := []string{
		metric.TagAsString("user_context", requestContext.UserContext.String()),
		metric.TagAsString("feed_type", enum.FeedTypeCrossSell.String()),
		metric.TagAsString("feed_context", request.Data.FeedContext.String()),
		metric.TagAsString("tenant_context", enum.TenantContextOrganic.String()),
		metric.TagAsString("method", "FetchCrossSellWidget"),
	}

	iopConfigRequest := &cohortConf.IopConfigRequest{
		UserId:        requestContext.UserId,
		UserContext:   requestContext.UserContext,
		FeedType:      enum.FeedTypeCrossSell,
		FeedContext:   request.Data.FeedContext,
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

	return s.buildResponse(dagExecResp, configResponse.ConfigBundle.GenerateWidgetOnTheFly, configResponse, &metricTags)
}

func (s *StandardCrossSellWidgetImpl) executeDAG(dagExecutionRequest *dag.ExecutionRequest,
	metricTags []string) (*dag.ExecutionResponse, error) {
	defer metric.TimingWithStart("dag_execution_latency", time.Now(), metricTags)
	return executor.Instance().Execute(dagExecutionRequest)
}

func (s *StandardCrossSellWidgetImpl) buildExecutionRequestData(request *GetCrossSellWidgetRequest,
	meta *api.RequestContext, metricTags []string) (*starterComponent.ExecutionRequestData, *api.Error) {
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

	// Get the last catalog ID
	var catalogId *int
	if len(request.Data.ParentEntityIds) > 0 {
		catalogId = &request.Data.ParentEntityIds[len(request.Data.ParentEntityIds)-1]
	} else {
		metric.Incr("parent_entity_ids_empty_error", metricTags)
		log.Error().Msgf("ParentEntityIds cannot be empty for cross sell widget request - %#v", request)
		return nil, api.NewBadRequestError("ParentEntityIds cannot be empty")
	}

	return &starterComponent.ExecutionRequestData{
		UserId:   meta.UserId,
		FeedType: enum.FeedTypeCrossSell,
		FeedId: starterComponent.FeedId{
			CatalogId: catalogId,
			SSCatId:   &request.Data.SubSubCategoryId,
		},
		FeedMetaData:  request.Data.Meta,
		FeedContext:   request.Data.FeedContext,
		UserContext:   meta.UserContext,
		IopId:         uuid.New().String(),
		TenantContext: enum.TenantContextOrganic,
		Context:       context,
		EntityType:    request.Data.ParentEntityType,
	}, nil
}

func (s *StandardCrossSellWidgetImpl) buildResponse(resp *dag.ExecutionResponse, iopConfig *dagConfig.IOP,
	iopConfigResponse *cohortConf.IopConfigResponse, metricTags *[]string) (*GetCrossSellWidgetResponse,
	*api.Error) {

	// fetch result from the result component mentioned in the iop config
	res, err := resp.Results[iopConfig.Config.ResultComponent].(*result.Future).Get()
	if err != nil {
		metric.Incr("dag_result_fetch_error", *metricTags)
		log.Error().Err(err).Msgf("Error in getting result from result component - %s", iopConfig.Config.ResultComponent)
		return nil, api.NewInternalServerError(err.Error())
	}
	metric.Incr("dag_result_fetch_success", *metricTags)
	response := res.(*starterComponent.Response)

	// building candidate with meta information
	similarEntitiesResponses := make([]handler.SimilarCandidatesResponse, 0, len(response.Candidates))
	for _, newCandidate := range response.Candidates {
		cursor, ok := newCandidate.Meta.Context.Cursor()
		if !ok {
			cursor = ""
		}

		metaData, err := buildCrossSellWidgetMetaData(newCandidate)
		if err != nil {
			metric.Incr("build_cross_sell_widget_metadata_error", *metricTags)
			log.Error().Err(err).Msgf("Error in building cross sell widget metadata for candidate id - %d", newCandidate.Id)
			return nil, err.(*api.Error)
		}

		similarEntitiesResponses = append(similarEntitiesResponses, handler.SimilarCandidatesResponse{
			Id:         newCandidate.Id,
			Cursor:     cursor,
			TrackingId: newCandidate.Meta.TrackingId,
			MetaData:   metaData,
		})
	}

	tags := append(*metricTags, metric.TagAsString("widget_length", "cross_sell"))
	metric.Timing(metric.MethodLatency, time.Duration(len(similarEntitiesResponses))*time.Millisecond, tags)

	metric.Count("cross_sell_widget_count", int64(len(similarEntitiesResponses)), nil)

	if len(similarEntitiesResponses) == 0 {
		metric.Incr("cross_sell_widget_empty", *metricTags)
	}

	return &GetCrossSellWidgetResponse{
		handler.GetCrossSellWidgetResponse{
			Data: handler.ResponseData{
				TenantContext:   enum.TenantContextOrganic,
				HasNextEntity:   response.HasNext,
				SimilarEntities: similarEntitiesResponses,
			},
			Error: handler.Error{
				Message: "",
			},
		},
	}, nil
}

func buildCrossSellWidgetMetaData(candidate *starterComponent.Candidate) (handler.MetaData, error) {

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

	if groupId, ok := candidate.Meta.Context.GetString(groupIdKey); ok {
		context[groupIdKey] = groupId
	}

	sscatId, ok := candidate.Meta.Context.GetString(sscatIdKey)
	if !ok {
		log.Error().Msgf("sscatId is not found in candidate context, candidate id - %d", candidate.Id)
		return handler.MetaData{}, api.NewInternalServerError("sscatId is not found in candidate context")
	}
	context[sscatIdKey] = sscatId

	sscatName, ok := candidate.Meta.Context.GetString(sscatNameKey)
	if !ok {
		log.Error().Msgf("sscatName is not found in candidate context, candidate id - %d", candidate.Id)
		return handler.MetaData{}, api.NewInternalServerError("sscatName is not found in candidate context")
	}
	context[sscatNameKey] = sscatName

	productId, ok := candidate.Meta.Context.GetString(productIdKey)
	if !ok {
		log.Error().Msgf("productId is not found in candidate context, candidate id - %d", candidate.Id)
		return handler.MetaData{}, api.NewInternalServerError("productId is not found in candidate context")
	}
	context[productIdKey] = productId

	parentProductId, ok := candidate.Meta.Context.GetString(parentProductIdKey)
	if !ok {
		log.Error().Msgf("parentProductId is not found in candidate context, candidate id - %d", candidate.Id)
		return handler.MetaData{}, api.NewInternalServerError("parentProductId is not found in candidate context")
	}
	context[parentProductIdKey] = parentProductId

	parentCatalogId, ok := candidate.Meta.Context.GetString(parentCatalogIdKey)
	if !ok {
		log.Error().Msgf("parentCatalogId is not found in candidate context, candidate id - %d", candidate.Id)
		return handler.MetaData{}, api.NewInternalServerError("parentCatalogId is not found in candidate context")
	}
	context[parentCatalogIdKey] = parentCatalogId

	return handler.MetaData{
		Scores:    scores,
		Source:    candidate.Meta.Source,
		Context:   context,
		SubTenant: subTenant,
	}, nil
}
