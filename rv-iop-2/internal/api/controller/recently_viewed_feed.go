package controller

import (
	"github.com/Meesho/go-core/api"
	"github.com/Meesho/rv-iop/internal/api/handler"
	"github.com/Meesho/rv-iop/internal/api/handler/ad"
	"github.com/Meesho/rv-iop/internal/api/handler/exploit"
	"github.com/Meesho/rv-iop/internal/api/handler/explore"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type RecentlyViewedFeedImpl interface {
	FetchRvSubstituteExploitFeed(context *gin.Context)
	FetchRvSubstituteExploreFeed(context *gin.Context)
	FetchRvSubstituteAdFeed(context *gin.Context)
}

type StandardRecentlyViewedFeedImpl struct {
	RvSubstituteFeedImpl   exploit.RvSubstituteFeedImpl
	RvSubstituteCtFeedImpl explore.RvSubstituteFeedImpl
	RvSubstituteAdFeedImpl ad.RvSubstituteFeedImpl
}

func NewStandardRecentlyViewedFeedImpl(exploitFeedImpl exploit.RvSubstituteFeedImpl, exploreFeedImpl explore.RvSubstituteFeedImpl, adFeedImpl ad.RvSubstituteFeedImpl) *StandardRecentlyViewedFeedImpl {
	if exploitFeedImpl == nil {
		log.Panic().Msgf("exploit feed controller interface cannot be nil")
	}
	if exploreFeedImpl == nil {
		log.Panic().Msgf("explore feed controller interface cannot be nil")
	}
	if adFeedImpl == nil {
		log.Panic().Msgf("ad feed controller interface cannot be nil")
	}

	return &StandardRecentlyViewedFeedImpl{
		RvSubstituteFeedImpl:   exploitFeedImpl,
		RvSubstituteCtFeedImpl: exploreFeedImpl,
		RvSubstituteAdFeedImpl: adFeedImpl,
	}
}

// FetchRvSubstituteAdFeed  pass the request to service layer and gives similar rvSubstitute feed response
// as output
func (s *StandardRecentlyViewedFeedImpl) FetchRvSubstituteAdFeed(context *gin.Context) {
	var request ad.GetRvSubstituteAdFeedRequest
	if err := context.BindJSON(&request); err != nil {
		log.Error().Err(err).Msg("Error in binding request body")
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	// check for valid request
	if ok, err := handler.IsValidGetRecentlyViewedFeedRequest(request.GetRecentlyViewedFeedRequest); !ok {
		log.Error().Err(err).Msgf("invalid request - %#v", request)
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	requestContext, err := api.GetRequestContext(context)
	if err != nil {
		log.Error().Err(err).Msgf("Error in getting request context, headers - %#v", context.Request.Header)
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	resp, apiError := s.RvSubstituteAdFeedImpl.FetchRvSubstituteAd(&request, requestContext)
	if apiError != nil {
		_ = context.Error(apiError)
		return
	}
	context.JSON(200, resp)
}

// FetchRvSubstituteExploitFeed pass the request to service layer and gives the similar exploit feed response
// as output.
func (s *StandardRecentlyViewedFeedImpl) FetchRvSubstituteExploitFeed(context *gin.Context) {
	var request exploit.GetRvSubstituteOrganicFeedRequest
	if err := context.BindJSON(&request); err != nil {
		log.Error().Err(err).Msg("Error in binding request body")
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	// check for valid request
	if ok, err := handler.IsValidGetRecentlyViewedFeedRequest(request.GetRecentlyViewedFeedRequest); !ok {
		log.Error().Err(err).Msgf("invalid request - %#v", request)
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	requestContext, err := api.GetRequestContext(context)
	if err != nil {
		log.Error().Err(err).Msgf("Error in getting request context, headers - %#v", context.Request.Header)
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	resp, apiError := s.RvSubstituteFeedImpl.FetchRvSubstitute(&request, requestContext)
	if apiError != nil {
		_ = context.Error(apiError)
		return
	}
	context.JSON(200, resp)
}

func (s *StandardRecentlyViewedFeedImpl) FetchRvSubstituteExploreFeed(context *gin.Context) {
	var request explore.GetRvSubstituteCtFeedRequest
	if err := context.BindJSON(&request); err != nil {
		log.Error().Err(err).Msg("Error in binding request body")
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	// check for valid request
	if ok, err := handler.IsValidGetRecentlyViewedFeedRequest(request.GetRecentlyViewedFeedRequest); !ok {
		log.Error().Err(err).Msgf("invalid request - %#v", request)
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	requestContext, err := api.GetRequestContext(context)
	if err != nil {
		log.Error().Err(err).Msgf("Error in getting request context, headers - %#v", context.Request.Header)
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}
	resp, apiError := s.RvSubstituteCtFeedImpl.FetchRvSubstituteCt(&request, requestContext)
	if apiError != nil {
		_ = context.Error(apiError)
		return
	}
	context.JSON(200, resp)
}
