package redisdb

import (
	"context"
	"fmt"
	redis "github.com/redis/go-redis/v9"
	"os"
	"time"
)

type RedisClient struct {
	client *redis.Client
}

func NewClient(db int) *RedisClient {
	//ctx := context.Background()

	uri := os.Getenv("REDIS_URI")
	rdb := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: "", // no password set
		DB:       db, // use default DB
	})

	return &RedisClient{
		client: rdb,
	}
}

func (r *RedisClient) Set(key string, value interface{}) error {
	ctx := context.Background()
	err := r.client.Set(ctx, key, value, 0).Err()
	return err
}

func (r *RedisClient) SetWithExpiration(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	err := r.client.Set(ctx, key, value, 0).Err()
	return err
}

func (r *RedisClient) Get(key string) (string, error) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()

	if err == redis.Nil {
		//fmt.Println("does not exist", key)
		return val, err
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println(key, val)
		return val, err
	}
}

func (r *RedisClient) GetAs(key string, getType interface{}) (any, error) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()

	var val2 any = nil

	switch t := getType.(type) {
	default:
		fmt.Printf("unexpected type %T", t) // %T prints whatever type t has
		err = fmt.Errorf("unexpected type %T", t)
	case bool:
		fmt.Printf("boolean %t\n", t) // t has type bool
		val2 = getType.(bool)
	case int:
		fmt.Printf("integer %d\n", t) // t has type int
		val2 = getType.(int)
	case string:
		fmt.Printf("string %s\n", t) // t has type int
		val2 = getType.(string)
	case *bool:
		fmt.Printf("pointer to boolean %t\n", *t) // t has type *bool
		val2 = getType.(*bool)
	case *int:
		fmt.Printf("pointer to integer %d\n", *t) // t has type *int
		val2 = getType.(*int)
	case *string:
		fmt.Printf("pointer to string %s\n", t) // t has type int
		val2 = getType.(*string)
	}

	if err == redis.Nil {
		fmt.Println("key2 does not exist")
		return val2, err
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val)
		return val2, err
	}
}

func (r *RedisClient) Del(key string) (int64, error) {
	ctx := context.Background()
	val, err := r.client.Del(ctx, key).Result()

	if err == redis.Nil {
		//fmt.Println("does not exist", key)
		return val, err
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println(key, val)
		return val, err
	}
}

//
//err := rdb.Set(ctx, "key", "value", 0).Err()
//if err != nil {
//panic(err)
//}
//
//val, err := rdb.Get(ctx, "key").Result()
//if err != nil {
//panic(err)
//}
//fmt.Println("key", val)
//
//val2, err := rdb.Get(ctx, "key2").Result()
//if err == redis.Nil {
//fmt.Println("key2 does not exist")
//} else if err != nil {
//panic(err)
//} else {
//fmt.Println("key2", val2)
//}
//// Output: key value
//// key2 does not exist
