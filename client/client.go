package client

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type Client interface {
	FetchPdpFeed(request *PdpFeedRequest) (response *PdpFeedResponse, err error)
}

func setHeadersInContext(Headers map[string]string, ctx context.Context) context.Context {
	md := metadata.New(Headers)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}
