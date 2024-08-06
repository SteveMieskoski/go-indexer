package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
	base "src/utils"
)

type AccessList struct {
	Address     base.Address `json:"address"`
	StorageKeys []Hash       `json:"storageKeys,omitempty"`
}

type MongoAccessList struct {
	Address     string   `bson:"address" json:"address"`
	StorageKeys []string `bson:"storageKeys,omitempty" json:"storageKeys,omitempty"`
}

func (s AccessList) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s AccessList) MongoFromGoType(lg AccessList) MongoAccessList {

	var convertedTopics []string
	for _, t := range lg.StorageKeys {
		convertedTopics = append(convertedTopics, t.String())
	}

	return MongoAccessList{
		Address:     lg.Address.String(),
		StorageKeys: convertedTopics,
	}
}

func (s AccessList) ProtobufFromGoType(lg AccessList) protobufLocal.AccessList {
	logString := s.MongoFromGoType(lg)

	return protobufLocal.AccessList{
		Address:     logString.Address,
		StorageKeys: logString.StorageKeys,
	}
}

func (s AccessList) MongoFromProtobufType(lg protobufLocal.AccessList) *MongoAccessList {

	return &MongoAccessList{
		Address:     lg.Address,
		StorageKeys: lg.StorageKeys,
	}
}

func (s AccessList) ProtobufFromMongoType(lg MongoAccessList) *protobufLocal.AccessList {

	return &protobufLocal.AccessList{
		Address:     lg.Address,
		StorageKeys: lg.StorageKeys,
	}
}
