package configModels

type RedisConnection struct {
	RedisMaxConnections     int    `json:"redisMaxConnections"`
	RedisMaxIdleConnections int    `json:"redisMaxIdleConnections"`
	RedisCon                string `json:"redisCon"`
	RedisAddress            string `json:"redisAddress"`
	PrintRedis              bool   `json:"printRedis"`
}
