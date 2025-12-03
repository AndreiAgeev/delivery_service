package redis

import (
	"context"
	"time"
)

// RedisClientInterface - интерфейс, реализующий часть методов redis.Client, необходимых для хендлеров
type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Hit()
	Miss()
}
