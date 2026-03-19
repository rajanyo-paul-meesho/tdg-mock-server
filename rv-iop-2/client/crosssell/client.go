package crosssell

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Meesho/go-core/circuitbreaker"
	"github.com/Meesho/go-core/grpcclient"
	coreUtils "github.com/Meesho/go-core/utils"
	grpc "github.com/Meesho/rv-iop/client/grpc/rv-iop"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

type CrossSell struct {
	crossSellClient grpc.RvIopServiceClient
	deadLine        int64
	adapter         Adapter
	CB              circuitbreaker.CircuitBreaker[*grpc.GetCrossSellWidgetRequest, *grpc.GetCrossSellWidgetResponse]
}

type Client interface {
	GetCrossSellWidget(request *CrossSellRequestData) (response *CrossSellResponseData, err error)
}

var (
	crossSellClientInstance Client
	crossSellClientOnce     sync.Once
)

func setHeadersInContext(headers map[string]string, ctx context.Context) context.Context {
	md := metadata.New(headers)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}

func getCrossSellCB(envPrefix string) circuitbreaker.CircuitBreaker[*grpc.GetCrossSellWidgetRequest, *grpc.GetCrossSellWidgetResponse] {
	return circuitbreaker.GetCircuitBreaker[*grpc.GetCrossSellWidgetRequest, *grpc.GetCrossSellWidgetResponse](circuitbreaker.BuildConfig(envPrefix))
}

func GetCrossSellClient(config *grpcclient.Config, envPrefix string) Client {
	crossSellClientOnce.Do(func() {
		conn := grpcclient.NewConnFromConfig(config, envPrefix)
		crossSellClientInstance = &CrossSell{
			crossSellClient: grpc.NewRvIopServiceClient(conn),
			deadLine:        conn.DeadLine,
			adapter:         Adapter{},
			CB:              getCrossSellCB(envPrefix),
		}
	})
	return crossSellClientInstance
}

func (c *CrossSell) GetCrossSellWidget(request *CrossSellRequestData) (response *CrossSellResponseData, err error) {

	reqProto, err := c.adapter.mapToCrossSellWidgetRequestProto(request)
	if err != nil {
		log.Error().Msgf("error mapping request to proto for ParentEntityIds %v", request.ParentEntityIds)
		return nil, err
	}

	res, err := c.contactServer(reqProto, c.deadLine, request.Headers)
	if err != nil {
		return nil, err
	}

	return c.adapter.mapCrossSellWidgetResponseFromProto(res), nil
}

func (c *CrossSell) contactServer(req *grpc.GetCrossSellWidgetRequest, deadline int64, headers map[string]string) (*grpc.GetCrossSellWidgetResponse, error) {
	timeout := time.Duration(deadline) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctxWithMetaData := setHeadersInContext(headers, ctx)
	defer cancel()

	resProto, err := c.CB.ExecuteForGrpc(ctxWithMetaData, req, c.crossSellClient.FetchCrossSellWidget)

	return c.processResponseProto(resProto, err, req)
}

func (c *CrossSell) processResponseProto(resProto *grpc.GetCrossSellWidgetResponse, err error, req *grpc.GetCrossSellWidgetRequest) (*grpc.GetCrossSellWidgetResponse, error) {
	if err != nil {
		return nil, err
	} else if coreUtils.IsNilPointer(resProto) {
		return nil, fmt.Errorf("found nil RV-IOP response for request %v", req)
	} else {
		return resProto, nil
	}
}
