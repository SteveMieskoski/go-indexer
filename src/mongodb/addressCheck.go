package mongodb

import (
	"bytes"
	"github.com/bits-and-blooms/bloom/v3"
	"src/redisdb"
)

type CheckAddress struct {
	redis  redisdb.RedisClient
	filter *bloom.BloomFilter
}

func NewAddressChecker() CheckAddress {
	redisClient := redisdb.NewClient(0)
	var filter *bloom.BloomFilter

	val, err := redisClient.Get("addressFilter")
	if err != nil {
		filter = bloom.NewWithEstimates(1000000, 0.01)
		var buf bytes.Buffer
		_, err := filter.WriteTo(&buf)
		if err != nil {
			panic(err)
		}
		err = redisClient.Set("addressFilter", buf.String())
		if err != nil {
			panic(err)
		}
	} else {
		filter = bloom.NewWithEstimates(1000000, 0.01)
		buf := new(bytes.Buffer)
		buf.WriteString(val)
		_, err := filter.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
	}

	return CheckAddress{
		redis:  *redisClient,
		filter: filter,
	}

}

func (c *CheckAddress) Exist(address string) bool {
	if c.filter.Test([]byte(address)) {
		return true
	} else {
		c.filter.Add([]byte(address))
		go func() {
			c.updateRedis()
		}()
		return false
	}
}

func (c *CheckAddress) updateRedis() {
	var buf bytes.Buffer
	_, err := c.filter.WriteTo(&buf)
	if err != nil {
		panic(err)
	}
	err = c.redis.Set("addressFilter", buf.String())
	if err != nil {
		panic(err)
	}
}
