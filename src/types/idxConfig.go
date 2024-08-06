package types

import "encoding/json"

type IdxConfigStruct struct {
	ClearKafka    bool `json:"clear_kafka"`
	ClearPostgres bool `json:"clear_postgres"`
	ClearRedis    bool `json:"clear_redis"`
	ClearConsumer bool `json:"clear_consumer"`
	DisableBeacon bool `json:"disable_beacon"`
	RunAsProducer bool `json:"run_as_producer"`
}

func (s IdxConfigStruct) String() string {
	bytes, _ := json.MarshalIndent(s, "", "   ")
	return string(bytes)
}
