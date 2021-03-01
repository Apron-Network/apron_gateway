package network_struct

type NewApiKeyRequest struct {
	ServiceId string
}

type ListApiKeysRequest struct {
	ServiceId string
	Start int
	Count int
}
