package explore

import (
	dagConfig "github.com/Meesho/dag-engine/v2/handlers/config"
	"github.com/Meesho/dag-engine/v2/handlers/dag"
	"github.com/Meesho/dag-engine/v2/handlers/dag/component/result"
	"github.com/Meesho/dag-engine/v2/handlers/dag/executor"
	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
	"github.com/Meesho/go-core/api"
	"github.com/Meesho/go-core/api/http"
	_ "github.com/Meesho/go-core/api/http"
	"github.com/Meesho/go-core/metric"
	starterComponent "github.com/Meesho/iop-component-starter/component"
	cohortConf "github.com/Meesho/iop-starter/cohort/config"
	"github.com/Meesho/rv-iop/internal/api/handler"
	"github.com/Meesho/rv-iop/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

type RvSubstituteFeedImpl interface {
	FetchRvSubstituteCt(request *GetRvSubstituteCtFeedRequest, requestContext *api.RequestContext) (
		*GetRvSubstituteCtFeedResponse, *api.Error)
}

// StandardRvSubstituteFeedImpl is the standard implementation of RvSubstituteFeedImpl
type StandardRvSubstituteFeedImpl struct {
	ServiceConf      *config.Service
	IopConfigHandler cohortConf.IopConfigHandler
}

func NewStandardRvSubstituteFeedImpl(serviceConf *config.Service,
	iopConfigHandler cohortConf.IopConfigHandler) *StandardRvSubstituteFeedImpl {
	if serviceConf == nil {
		log.Panic().Msgf("service conf cannot be nil")
	}
	if iopConfigHandler == nil {
		log.Panic().Msgf("iop config handler cannot be nil")
	}
	return &StandardRvSubstituteFeedImpl{
		ServiceConf:      serviceConf,
		IopConfigHandler: iopConfigHandler,
	}
}

func (s *StandardRvSubstituteFeedImpl) FetchRvSubstituteCt(request *GetRvSubstituteCtFeedRequest,
	requestContext *api.RequestContext) (*GetRvSubstituteCtFeedResponse, *api.Error) {
	metricTags := []string{
		metric.TagAsString("user_context", requestContext.UserContext.String()),
		metric.TagAsString("feed_type", enum.FeedTypeRecentlyViewedCatalogRecommendation.String()),
		metric.TagAsString("feed_context", request.Data.FeedContext.String()),
		metric.TagAsString("tenant_context", enum.TenantContextCT.String()),
		metric.TagAsString("method", "FetchRvSubstitute"),
	}

	iopConfigRequest := &cohortConf.IopConfigRequest{
		UserId:        requestContext.UserId,
		UserContext:   requestContext.UserContext,
		FeedType:      enum.FeedTypeRecentlyViewedCatalogRecommendation,
		FeedContext:   request.Data.FeedContext,
		TenantContext: enum.TenantContextCT,
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
	dagExecResp, err := s.executeDAG(&dag.ExecutionRequest{
		IOPConfig: iopConfig,
		Data:      s.buildExecutionRequestData(request, requestContext),
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
		log.Error().Err(err).Msgf("Error in getting executing dag request - %#v", iopConfigRequest)
		return nil, api.NewInternalServerError(err.Error())
	}
	metric.Incr("dag_execution_success", metricTags)

	return buildResponse(request, dagExecResp, configResponse.ConfigBundle.GenerateFeedOnTheFly, configResponse, &metricTags)
}

func (s *StandardRvSubstituteFeedImpl) emptyResponse() (*GetRvSubstituteCtFeedResponse, *api.Error) {
	return &GetRvSubstituteCtFeedResponse{
		handler.GetRecentlyViewedFeedResponse{
			Data: handler.ResponseData{
				TenantContext:   enum.TenantContextCT,
				HasNextEntity:   false,
				SimilarEntities: make([]handler.SimilarCandidatesResponse, 0),
			},
			Error: handler.Error{
				Message: "",
			},
		},
	}, nil

}

func (s *StandardRvSubstituteFeedImpl) executeDAG(dagExecutionRequest *dag.ExecutionRequest,
	metricTags []string) (*dag.ExecutionResponse, error) {
	defer metric.TimingWithStart("dag_execution_latency", time.Now(), metricTags)
	return executor.Instance().Execute(dagExecutionRequest)
}

func (s *StandardRvSubstituteFeedImpl) buildExecutionRequestData(request *GetRvSubstituteCtFeedRequest,
	meta *api.RequestContext) *starterComponent.ExecutionRequestData {
	context := starterComponent.NewRequestContext()
	if len(request.Data.Cursor) != 0 {
		context.SetCursor(request.Data.Cursor)
	}
	context.SetRequestedLimit(request.Data.Limit)
	context.Set(http.HeaderMeeshoUserStateCode, meta.UserStateCode)
	return &starterComponent.ExecutionRequestData{
		UserId:   meta.UserId,
		FeedType: enum.FeedTypeRecentlyViewedCatalogRecommendation,
		FeedId: starterComponent.FeedId{
			SSCatId: &request.Data.SubSubCategoryId,
		},
		FeedContext:   request.Data.FeedContext,
		UserContext:   meta.UserContext,
		IopId:         uuid.New().String(),
		TenantContext: enum.TenantContextCT,
		Context:       context,
		EntityType:    request.Data.ParentEntityType,
	}

}
func buildResponse(request *GetRvSubstituteCtFeedRequest, resp *dag.ExecutionResponse, iopConfig *dagConfig.IOP,
	iopConfigResponse *cohortConf.IopConfigResponse, metricTags *[]string) (*GetRvSubstituteCtFeedResponse,
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
	meta := buildMeta(request)
	for _, newCandidate := range response.Candidates {
		cursor, ok := newCandidate.Meta.Context.Cursor()
		if !ok {
			log.Panic().Msgf("cursor is not set for candidates, dag name - %s", iopConfigResponse.ConfigVariantName)
		}
		similarEntitiesResponses = append(similarEntitiesResponses, handler.SimilarCandidatesResponse{
			Id:         newCandidate.Id,
			Cursor:     cursor,
			TrackingId: newCandidate.Meta.TrackingId,
			Meta:       meta,
		})
	}
	return &GetRvSubstituteCtFeedResponse{
		handler.GetRecentlyViewedFeedResponse{
			Data: handler.ResponseData{
				TenantContext:   enum.TenantContextCT,
				HasNextEntity:   response.HasNext,
				SimilarEntities: similarEntitiesResponses,
			},
			Error: handler.Error{
				Message: "",
			},
		},
	}, nil

}

// buildMeta it builds metadata from the request
func buildMeta(request *GetRvSubstituteCtFeedRequest) map[string]string {
	return map[string]string{
		"pei": strconv.Itoa(request.Data.SubSubCategoryId),
		"pet": request.Data.ParentEntityType.String(),
	}
}
