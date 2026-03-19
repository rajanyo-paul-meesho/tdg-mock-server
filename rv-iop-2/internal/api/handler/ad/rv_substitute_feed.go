package ad

import (
	dagConfig "github.com/Meesho/dag-engine/v2/handlers/config"
	"github.com/Meesho/dag-engine/v2/handlers/dag"
	"github.com/Meesho/dag-engine/v2/handlers/dag/component/result"
	"github.com/Meesho/dag-engine/v2/handlers/dag/executor"
	"github.com/Meesho/feed-commons-go/v2/pkg/data"
	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
	"github.com/Meesho/go-core/api"
	"github.com/Meesho/go-core/api/http"
	_ "github.com/Meesho/go-core/api/http"
	"github.com/Meesho/go-core/metric"
	starterComponent "github.com/Meesho/iop-component-starter/component"
	data2 "github.com/Meesho/iop-component-starter/data"
	cohortConf "github.com/Meesho/iop-starter/cohort/config"
	"github.com/Meesho/rv-iop/internal/api/handler"
	"github.com/Meesho/rv-iop/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

const (
	cpcScoreKey        = "cpc_score"
	rankingScoreKey    = "default"
	similarityScoreKey = "score"
)

type RvSubstituteFeedImpl interface {
	FetchRvSubstituteAd(request *GetRvSubstituteAdFeedRequest, requestContext *api.RequestContext) (
		*GetRvSubstituteAdFeedResponse, *api.Error)
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
func (s *StandardRvSubstituteFeedImpl) FetchRvSubstituteAd(request *GetRvSubstituteAdFeedRequest,
	requestContext *api.RequestContext) (*GetRvSubstituteAdFeedResponse, *api.Error) {
	metricTags := []string{
		metric.TagAsString("user_context", requestContext.UserContext.String()),
		metric.TagAsString("feed_type", enum.FeedTypeRecentlyViewedCatalogRecommendation.String()),
		metric.TagAsString("feed_context", request.Data.FeedContext.String()),
		metric.TagAsString("tenant_context", enum.TenantContextAd.String()),
		metric.TagAsString("method", "FetchRvSubstitute"),
	}

	iopConfigRequest := &cohortConf.IopConfigRequest{
		UserId:        requestContext.UserId,
		UserContext:   requestContext.UserContext,
		FeedType:      enum.FeedTypeRecentlyViewedCatalogRecommendation,
		FeedContext:   request.Data.FeedContext,
		TenantContext: enum.TenantContextAd,
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

	return buildResponse(dagExecResp, configResponse.ConfigBundle.GenerateFeedOnTheFly, configResponse, &metricTags)
}

func (s *StandardRvSubstituteFeedImpl) emptyResponse() (*GetRvSubstituteAdFeedResponse, *api.Error) {
	return &GetRvSubstituteAdFeedResponse{
		handler.GetRecentlyViewedFeedResponse{
			Data: handler.ResponseData{
				TenantContext:   enum.TenantContextAd,
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

func (s *StandardRvSubstituteFeedImpl) buildExecutionRequestData(request *GetRvSubstituteAdFeedRequest,
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
		TenantContext: enum.TenantContextAd,
		Context:       context,
		EntityType:    request.Data.ParentEntityType,
	}

}

func buildResponse(resp *dag.ExecutionResponse, iopConfig *dagConfig.IOP,
	iopConfigResponse *cohortConf.IopConfigResponse, metricTags *[]string) (*GetRvSubstituteAdFeedResponse,
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
			log.Panic().Msgf("cursor is not set for candidates, dag name - %s", iopConfigResponse.ConfigVariantName)
		}
		var metaData *handler.MetaData
		metaData, err = buildMetaData(newCandidate)
		if err != nil {
			log.Error().Err(err).Msgf("error in building meta data for catalog id %v, skipping this candidate", newCandidate.Id)
			metric.Incr("build_meta_data_failures", *metricTags)
			continue
		}
		similarEntitiesResponses = append(similarEntitiesResponses, handler.SimilarCandidatesResponse{
			Id:         newCandidate.Id,
			Cursor:     cursor,
			TrackingId: newCandidate.Meta.TrackingId,
			MetaData:   *metaData,
		})
	}
	return &GetRvSubstituteAdFeedResponse{
		handler.GetRecentlyViewedFeedResponse{
			Data: handler.ResponseData{
				TenantContext:   enum.TenantContextAd,
				HasNextEntity:   response.HasNext,
				SimilarEntities: similarEntitiesResponses,
				Slots:           getSlotsAndPushMetrics(response.FeedMetaData, metricTags),
			},
			Error: handler.Error{
				Message: "",
			},
		},
	}, nil

}

func buildMetaData(candidate *starterComponent.Candidate) (*handler.MetaData, error) {
	subTenant, ok := candidate.Meta.Context.GetString(data.SubTenantKey)
	if !ok {
		return &handler.MetaData{}, nil
	}
	// build sse (server side events)
	sse := data.NewSse()
	cpcScore := parseScore(&candidate.Meta.Scores, cpcScoreKey)
	rankingScore := parseScore(&candidate.Meta.Scores, rankingScoreKey)
	similarityScore := parseScore(&candidate.Meta.Scores, similarityScoreKey)

	sse[data.CatalogBasedFeedTypeKey] = subTenant
	sse[data.FeedSourceKey] = data.FeedSourcePersonalized
	sse[data.SlotConfigSourceKey] = data.SlotConfigSourceIop
	sse[data.CgAlgoKey], _ = candidate.Meta.Context.GetString(data.SourceCg)
	sse[data.CpcKey] = strconv.FormatFloat(cpcScore, 'f', -1, 64)
	sse[data.FeedSortScoreKey] = strconv.FormatFloat(rankingScore, 'f', -1, 64)
	sse[data.ScoreKey] = strconv.FormatFloat(similarityScore, 'f', -1, 64)
	sse[data.SimilarityScoreKey] = strconv.FormatFloat(similarityScore, 'f', -1, 64)

	return &handler.MetaData{
		SubTenant: subTenant,
		Sse:       sse,
		Scores:    candidate.Meta.Scores,
	}, nil
}

func getSlotsAndPushMetrics(feedMetaData map[string]interface{}, metricTags *[]string) []int32 {
	if feedMetaData == nil {
		metric.Incr("slots_not_found", *metricTags)
		log.Debug().Msgf("FeedMetaData is nil in dag response for %#v tenant", enum.TenantContextAd.String())
		return nil
	}
	if feedMetaData[data.SlotsKey] == nil {
		metric.Incr("feed_metadata_not_found", *metricTags)
		log.Debug().Msgf("slots is nil in dag response for %#v tenant", enum.TenantContextAd.String())
		return nil
	}
	slots, ok := feedMetaData[data.SlotsKey].([]int32)
	if !ok {
		metric.Incr("slots_parsing_failed", *metricTags)
		log.Error().Msgf("failure in parsing slots in dag response for %#v tenant", enum.TenantContextAd.String())
		return make([]int32, 0)
	}
	return slots
}

func parseScore(scores *data2.Scores, key string) float64 {
	scoreStr, exists := scores.Get(key)
	if !exists || len(scoreStr) == 0 {
		return 0.00
	}

	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		log.Error().Err(err).Msgf("Error while converting string to float64 of %s: %v", key, score)
		return 0.00
	}
	return score
}
