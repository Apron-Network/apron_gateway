package models

func (k *ApronApiKey) StoreBucketName() string {
	return ServiceApiKeyStorageBucketName(k.ServiceId)
}

