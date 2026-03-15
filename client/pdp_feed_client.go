package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	grpc "github.com/Meesho/feed-aggregator-go/client/grpc/pdp"
	"github.com/Meesho/go-core/grpcclient"
	coreUtils "github.com/Meesho/go-core/utils"
	"github.com/rs/zerolog/log"
)

type PdpFeedClient struct {
	pdpClient grpc.PdpFeedHandlerClient
	deadLine  int64
	adapter   Adapter
}

var (
	pdpFeedClientInstance Client
	pdpFeedClientOnce     sync.Once
)

func GetPdpFeedClient(config *grpcclient.Config, envPrefix string) Client {
	pdpFeedClientOnce.Do(func() {
		conn := grpcclient.NewConnFromConfig(config, envPrefix)
		pdpFeedClientInstance = &PdpFeedClient{
			pdpClient: grpc.NewPdpFeedHandlerClient(conn),
			deadLine:  conn.DeadLine,
			adapter:   Adapter{},
		}
	})
	return pdpFeedClientInstance
}

func (c *PdpFeedClient) FetchPdpFeed(request *PdpFeedRequest) (response *PdpFeedResponse, err error) {
	// Map the request to proto
	reqProto, err := c.adapter.mapToPdpFeedRequestProto(request)
	if err != nil {
		log.Error().Msgf("error mapping request to proto for CatalogId %v", request.CatalogId)
		return nil, err
	}

	// Contact feed-aggregator gRPC server
	res, err := c.contactServer(reqProto, c.deadLine, request.Headers)
	if err != nil {
		return nil, err
	}

	// Map the proto to response
	return c.adapter.mapFromPdpFeedResponseProto(res), nil
}

// contactServer contacts feed-aggregator gRPC server
func (c *PdpFeedClient) contactServer(req *grpc.RecommendationsRequest, deadline int64, Headers map[string]string) (*grpc.RecommendationResponse, error) {
	// Set the deadline for the request
	timeout := time.Duration(deadline) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctxWithMetaData := setHeadersInContext(Headers, ctx)
	defer cancel()

	resProto, err := c.pdpClient.FetchPdpFeed(ctxWithMetaData, req)
	return c.processResponseProto(resProto, err, req)
}

func (c *PdpFeedClient) processResponseProto(resProto *grpc.RecommendationResponse, err error, req *grpc.RecommendationsRequest) (*grpc.RecommendationResponse, error) {
	if err != nil {
		log.Error().Msgf("Error while calling feed-aggregator gRPC server: %v", err)
		return nil, err
	}

	if resProto == nil {
		log.Error().Msgf("Received nil response from feed-aggregator gRPC server for CatalogId: %v", req.CatalogId)
		return nil, errors.New("received nil response from feed-aggregator gRPC server")
	}

	if !coreUtils.IsEmptyString(resProto.Error) {
		log.Error().Msgf("Received error from feed-aggregator gRPC server: %v for CatalogId: %v", resProto.Error, req.CatalogId)
		return nil, fmt.Errorf("feed-aggregator gRPC server error: %s", resProto.Error)
	}

	log.Info().Msgf("Successfully received response from feed-aggregator gRPC server for CatalogId: %v, catalog count: %v", req.CatalogId, len(resProto.Catalogs))
	return resProto, nil
}
