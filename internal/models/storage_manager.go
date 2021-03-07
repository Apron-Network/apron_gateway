package models

import (
	"errors"

	"github.com/go-redis/redis/v8"

	"apron.network/gateway/internal"
)

type StorageManager struct {
	// TODO: more db support will be added
	RedisClient *redis.Client
}

func parseHscanResultToObjectMap(rcds []string) (map[string]string, uint, error) {
	if len(rcds)%2 != 0 {
		return nil, 0, errors.New("record length should be even number")
	}

	resultCount := uint(len(rcds) / 2)
	rslt := make(map[string]string)
	for idx := 0; idx < len(rcds); idx += 2 {
		rslt[rcds[idx]] = rcds[idx+1]
	}

	return rslt, resultCount, nil
}
func (s *StorageManager) IsKeyExisting(key string) bool {
	existing, err := s.RedisClient.Exists(internal.Ctx(), key).Result()
	internal.CheckError(err)
	return existing == 1
}

func (s *StorageManager) IsKeyExistingInBucket(table, keyValue string) bool {
	existing, err := s.RedisClient.HExists(internal.Ctx(), table, keyValue).Result()
	internal.CheckError(err)
	return existing
}

func (s *StorageManager) SaveBinaryKeyData(table, keyVal string, content []byte) error {
	_, err := s.RedisClient.HSet(internal.Ctx(), table, keyVal, content).Result()
	return err
}

func (s *StorageManager) DeleteKey(table, keyVal string) error {
	_, err := s.RedisClient.HDel(internal.Ctx(), table, keyVal).Result()
	return err
}

func (s *StorageManager) FetchRecords(table string, startIdx int, pattern string, size int) (map[string]string, uint64, uint, error) {
	rcds, cursor, err := s.RedisClient.HScan(internal.Ctx(), table, uint64(startIdx), pattern, int64(size)).Result()
	internal.CheckError(err)
	scanResultMap, resultCount, err := parseHscanResultToObjectMap(rcds)
	return scanResultMap, cursor, resultCount, err
}

func (s *StorageManager) GetRecord(table, key string) (string, error) {
	return s.RedisClient.HGet(internal.Ctx(), table, key).Result()
}
