package controller

import (
	"github.com/Meesho/go-core/api"
	rvinteraction "github.com/Meesho/rv-iop/internal/api/handler/rv-interaction"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type RvInteractionsImpl interface {
	FetchRvInteractionsFeed(context *gin.Context)
	FetchRvInteractionsWidget(context *gin.Context)
}

type StandardRvInteractionsImpl struct {
	RvInteractionsFeedHandler   rvinteraction.RvInteractionsFeedImpl
	RvInteractionsWidgetHandler rvinteraction.RvInteractionsWidgetImpl
}

func NewStandardRvInteractionsImpl(
	rvInteractionsFeedHandler rvinteraction.RvInteractionsFeedImpl,
	rvInteractionsWidgetHandler rvinteraction.RvInteractionsWidgetImpl) *StandardRvInteractionsImpl {
	if rvInteractionsFeedHandler == nil {
		log.Panic().Msgf("rv interactions feed handler cannot be nil")
	}
	if rvInteractionsWidgetHandler == nil {
		log.Panic().Msgf("rv interaction widget handler cannot be nil")
	}

	return &StandardRvInteractionsImpl{
		RvInteractionsFeedHandler:   rvInteractionsFeedHandler,
		RvInteractionsWidgetHandler: rvInteractionsWidgetHandler,
	}
}

// FetchRvInteractionsFeed pass the request to service layer and gives rv interactions feed response
// as output
func (s *StandardRvInteractionsImpl) FetchRvInteractionsFeed(context *gin.Context) {
	var request rvinteraction.GetRvInteractionsFeedRequest
	if err := context.BindJSON(&request); err != nil {
		log.Error().Err(err).Msg("Error in binding request body")
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}

	// Validate request
	if ok, err := rvinteraction.IsValidGetRvInteractionsFeedRequest(&request); !ok {
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

	resp, apiError := s.RvInteractionsFeedHandler.FetchRvInteractionsFeed(&request, requestContext)
	if apiError != nil {
		_ = context.Error(apiError)
		return
	}

	context.JSON(200, resp)
}

// FetchRvInteractionsWidget pass the request to service layer and gives rv interaction widget response
// as output
func (s *StandardRvInteractionsImpl) FetchRvInteractionsWidget(context *gin.Context) {
	var request rvinteraction.GetRvInteractionsWidgetRequest
	if err := context.BindJSON(&request); err != nil {
		log.Error().Err(err).Msg("Error in binding request body")
		_ = context.Error(api.NewBadRequestError(err.Error()))
		return
	}

	// Validate request
	if ok, err := rvinteraction.IsValidGetRvInteractionsWidgetRequest(&request); !ok {
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

	resp, apiError := s.RvInteractionsWidgetHandler.FetchRvInteractionsWidget(&request, requestContext)
	if apiError != nil {
		_ = context.Error(apiError)
		return
	}

	context.JSON(200, resp)
}
