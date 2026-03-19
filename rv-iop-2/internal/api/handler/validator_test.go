package handler

import (
	"testing"

	"github.com/Meesho/feed-commons-go/v2/pkg/enum"
	"github.com/stretchr/testify/assert"
)

func TestIsValidGetRecentlyViewedFeedRequest(t *testing.T) {
	tests := []struct {
		name    string
		request GetRecentlyViewedFeedRequest
		wantOk  bool
		wantErr string
	}{
		{
			name: "valid with positive limit and sscat",
			request: GetRecentlyViewedFeedRequest{
				Data: RequestData{
					Limit:            10,
					SubSubCategoryId: 100,
					FeedContext:      enum.FeedContextDefault,
				},
			},
			wantOk:  true,
			wantErr: "",
		},
		{
			name: "valid wishlist skips sscat check",
			request: GetRecentlyViewedFeedRequest{
				Data: RequestData{
					Limit:            5,
					SubSubCategoryId: 0,
					FeedContext:      enum.FeedContextWishlist,
				},
			},
			wantOk:  true,
			wantErr: "",
		},
		{
			name: "invalid when limit <= 0",
			request: GetRecentlyViewedFeedRequest{
				Data: RequestData{
					Limit:            0,
					SubSubCategoryId: 100,
					FeedContext:      enum.FeedContextDefault,
				},
			},
			wantOk:  false,
			wantErr: "limit should be positive",
		},
		{
			name: "invalid when sscat <= 0 and not wishlist",
			request: GetRecentlyViewedFeedRequest{
				Data: RequestData{
					Limit:            10,
					SubSubCategoryId: 0,
					FeedContext:      enum.FeedContextDefault,
				},
			},
			wantOk:  false,
			wantErr: "invalid sscat id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := IsValidGetRecentlyViewedFeedRequest(tt.request)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidGetCrossSellWidgetRequest(t *testing.T) {
	tests := []struct {
		name    string
		request GetCrossSellWidgetRequest
		wantOk  bool
		wantErr string
	}{
		{
			name: "valid",
			request: GetCrossSellWidgetRequest{
				Data: CrossSellRequestData{
					Limit:            10,
					ParentEntityIds:  []int{1, 2},
				},
			},
			wantOk:  true,
			wantErr: "",
		},
		{
			name: "invalid when limit <= 0",
			request: GetCrossSellWidgetRequest{
				Data: CrossSellRequestData{
					Limit:            0,
					ParentEntityIds:  []int{1},
				},
			},
			wantOk:  false,
			wantErr: "limit should be positive",
		},
		{
			name: "invalid when parent entity ids empty",
			request: GetCrossSellWidgetRequest{
				Data: CrossSellRequestData{
					Limit:           10,
					ParentEntityIds: nil,
				},
			},
			wantOk:  false,
			wantErr: "invalid parent entity id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := IsValidGetCrossSellWidgetRequest(tt.request)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidGetCrossSellFeedRequest(t *testing.T) {
	tests := []struct {
		name    string
		request GetCrossSellFeedRequest
		wantOk  bool
		wantErr string
	}{
		{
			name: "valid",
			request: GetCrossSellFeedRequest{
				Data: CrossSellRequestData{
					Limit:            10,
					ParentEntityIds:  []int{1, 2},
					SubSubCategoryId: 100,
				},
			},
			wantOk:  true,
			wantErr: "",
		},
		{
			name: "invalid when limit <= 0",
			request: GetCrossSellFeedRequest{
				Data: CrossSellRequestData{
					Limit:            0,
					ParentEntityIds:  []int{1},
					SubSubCategoryId: 100,
				},
			},
			wantOk:  false,
			wantErr: "limit should be positive",
		},
		{
			name: "invalid when parent entity ids empty",
			request: GetCrossSellFeedRequest{
				Data: CrossSellRequestData{
					Limit:            10,
					ParentEntityIds: []int{},
					SubSubCategoryId: 100,
				},
			},
			wantOk:  false,
			wantErr: "invalid parent entity id",
		},
		{
			name: "invalid when sscat <= 0",
			request: GetCrossSellFeedRequest{
				Data: CrossSellRequestData{
					Limit:            10,
					ParentEntityIds:  []int{1},
					SubSubCategoryId: 0,
				},
			},
			wantOk:  false,
			wantErr: "invalid sscat id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := IsValidGetCrossSellFeedRequest(tt.request)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
