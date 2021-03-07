package handlers

import "apron.network/gateway/internal/models"

type ListApiKeysResponse struct {
	ServiceId  string
	Records    []models.ApronApiKey
	Count      uint
	NextCursor uint64
}
