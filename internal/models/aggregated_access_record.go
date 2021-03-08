package models

import (
	"encoding/json"
	"fmt"
	"time"

	"apron.network/gateway/internal"
)

type AggregatedAccessRecord struct {
	Id          uint64 `json:"id"`
	ServiceUuid string `json:"service_uuid"`
	UserKey     string `json:"user_key"`
	StartTime   uint64 `json:"start_time"`
	EndTime     uint64 `json:"end_time"`
	Usage       uint64 `json:"usage"`
	PricePlan   string `json:"price_plan"`
	cost        uint64 `json:"cost"`
}

func (r *AggregatedAccessRecord) Reset(startTime time.Time) {
	epochSecond := uint64(startTime.Unix())

	r.Id = epochSecond // TODO: Confirm how to generate the ID
	r.StartTime = epochSecond
	r.Usage = 0
}

func (r *AggregatedAccessRecord) ExportAndFlush() string {
	currentTime := time.Now().UTC()

	r.EndTime = uint64(currentTime.Unix())

	strData, err := json.Marshal(r)
	internal.CheckError(err)

	r.Reset(currentTime)

	return string(strData)
}

func AccessRecordStorageKeyFrom(serviceUuid, userKey string) string {
	return fmt.Sprintf("%s.%s", serviceUuid, userKey)
}
