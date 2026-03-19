package rvinteraction

import "errors"

func IsValidGetRvInteractionsFeedRequest(request *GetRvInteractionsFeedRequest) (bool, error) {
	if request.Data.UserId == "" {
		return false, errors.New("user_id cannot be empty")
	}

	if request.Data.SscatId <= 0 {
		return false, errors.New("sscat_id must be positive")
	}

	if request.Data.Limit <= 0 {
		return false, errors.New("limit must be positive")
	}

	return true, nil
}

func IsValidGetRvInteractionsWidgetRequest(request *GetRvInteractionsWidgetRequest) (bool, error) {
	if request.Data.UserId == "" {
		return false, errors.New("user_id cannot be empty")
	}

	if request.Data.Limit <= 0 {
		return false, errors.New("limit must be positive")
	}

	return true, nil
}
