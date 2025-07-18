// This generated by evm-cli, edit as necessary
package infras

import (
	"context"
	"fmt"

	"github.com/IlhamRobyana/user/configs"
	"github.com/go-redis/redis/v8"
)

// RedisNewClient create new instance of redis
func RedisNewClient(config configs.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Cache.Redis.Primary.Host, config.Cache.Redis.Primary.Port),
		Password: config.Cache.Redis.Primary.Password,
	})

	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(pong, err)

	return client
}
