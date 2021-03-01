package models

import "fmt"

func ServiceApiKeyStorageBucketName(service_id string) string {
	return fmt.Sprint("ApiKeyBucket:%s", service_id)
}