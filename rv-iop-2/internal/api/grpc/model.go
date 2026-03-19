package grpc

import (
	rv_iop "github.com/Meesho/rv-iop/client/grpc/rv-iop"
	"github.com/Meesho/rv-iop/internal/api/handler/ad"
	"github.com/Meesho/rv-iop/internal/api/handler/crosssell"
	"github.com/Meesho/rv-iop/internal/api/handler/exploit"
	"github.com/Meesho/rv-iop/internal/api/handler/explore"
	rvinteractionhandler "github.com/Meesho/rv-iop/internal/api/handler/rv-interaction"
)

type RvIopGrpcServer struct {
	rv_iop.UnimplementedRvIopServiceServer
	ExploitHandler              exploit.RvSubstituteFeedImpl
	ExploreHandler              explore.RvSubstituteFeedImpl
	AdHandler                   ad.RvSubstituteFeedImpl
	CrossSellWidgetHandler      crosssell.CrossSellWidgetImpl
	CrossSellFeedHandler        crosssell.CrossSellFeedImpl
	RvInteractionsFeedHandler   rvinteractionhandler.RvInteractionsFeedImpl
	RvInteractionsWidgetHandler rvinteractionhandler.RvInteractionsWidgetImpl
}

func NewRvIopGrpcServer(
	exploitHandler exploit.RvSubstituteFeedImpl,
	exploreHandler explore.RvSubstituteFeedImpl,
	adHandler ad.RvSubstituteFeedImpl,
	crossSellWidgetHandler crosssell.CrossSellWidgetImpl,
	crossSellFeedHandler crosssell.CrossSellFeedImpl,
	rvInteractionsFeedHandler rvinteractionhandler.RvInteractionsFeedImpl,
	rvInteractionsWidgetHandler rvinteractionhandler.RvInteractionsWidgetImpl,
) *RvIopGrpcServer {
	return &RvIopGrpcServer{
		ExploitHandler:              exploitHandler,
		ExploreHandler:              exploreHandler,
		AdHandler:                   adHandler,
		CrossSellWidgetHandler:      crossSellWidgetHandler,
		CrossSellFeedHandler:        crossSellFeedHandler,
		RvInteractionsFeedHandler:   rvInteractionsFeedHandler,
		RvInteractionsWidgetHandler: rvInteractionsWidgetHandler,
	}
}
