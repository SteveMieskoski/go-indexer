package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
	base "src/utils"
)

type AccessList struct {
	Address     base.Address `json:"address"`
	StorageKeys []base.Hash  `json:"storageKeys,omitempty"`
}

type MongoAccessList struct {
	Address     string   `bson:"address" json:"address"`
	StorageKeys []string `bson:"storageKeys,omitempty" json:"storageKeys,omitempty"`
}

func (s AccessList) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s AccessList) FromGoType(lg AccessList) MongoAccessList {

	var convertedTopics []string
	for _, t := range lg.StorageKeys {
		convertedTopics = append(convertedTopics, t.String())
	}

	return MongoAccessList{
		Address:     lg.Address.String(),
		StorageKeys: convertedTopics,
	}
}

func (s AccessList) ProtobufFromGoType(lg AccessList) protobufLocal.Block_Transaction_AccessList {
	logString := s.FromGoType(lg)

	return protobufLocal.Block_Transaction_AccessList{
		Address:     logString.Address,
		StorageKeys: logString.StorageKeys,
	}
}

func (s AccessList) MongoFromProtobufType(lg protobufLocal.Block_Transaction_AccessList) *MongoAccessList {

	return &MongoAccessList{
		Address:     lg.Address,
		StorageKeys: lg.StorageKeys,
	}
}

func (s AccessList) ProtobufFromMongoType(lg MongoAccessList) *protobufLocal.Block_Transaction_AccessList {

	return &protobufLocal.Block_Transaction_AccessList{
		Address:     lg.Address,
		StorageKeys: lg.StorageKeys,
	}
}
