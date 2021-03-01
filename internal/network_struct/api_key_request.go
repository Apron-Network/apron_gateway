package network_struct

import "apron.network/gateway/internal/models"

type NewApiKeyRequest struct {
	ServiceId string
}

type ListApiKeysRequest struct {
	ServiceId string
	Start     int
	Count     int
}
