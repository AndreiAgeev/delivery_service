package redis

import (
	"context"
	"delivery-system/internal/database"
	"delivery-system/internal/models"
	"encoding/json"
	"fmt"
	"time"

	"delivery-system/internal/config"
	"delivery-system/internal/logger"

	"github.com/go-redis/redis/v8"
)

// Client представляет клиент Redis
type Client struct {
	client  *redis.Client
	log     *logger.Logger
	metrics *RedisMetrics
}

// Connect создает подключение к Redis
func Connect(cfg *config.RedisConfig, log *logger.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Проверка подключения
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Successfully connected to Redis")

	return &Client{
		client:  rdb,
		log:     log,
		metrics: &RedisMetrics{},
	}, nil
}

// Close закрывает подключение к Redis
func (c *Client) Close() error {
	return c.client.Close()
}

// Set устанавливает значение с TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	c.log.WithField("key", key).Debug("Value set in Redis")
	return nil
}

// Get получает значение по ключу
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s not found", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	c.log.WithField("key", key).Debug("Value retrieved from Redis")
	return nil
}

// Delete удаляет значение по ключу
func (c *Client) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	c.log.WithField("key", key).Debug("Key deleted from Redis")
	return nil
}

// Exists проверяет существование ключа
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if key %s exists: %w", key, err)
	}

	return exists > 0, nil
}

// SetMultiple устанавливает несколько значений за одну операцию
func (c *Client) SetMultiple(ctx context.Context, values map[string]interface{}, ttl time.Duration) error {
	pipe := c.client.Pipeline()

	for key, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	c.log.WithField("count", len(values)).Debug("Multiple values set in Redis")
	return nil
}

// GetMultiple получает несколько значений за одну операцию
func (c *Client) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	result := make(map[string]string)
	for i, key := range keys {
		if values[i] != nil {
			result[key] = values[i].(string)
		}
	}

	c.log.WithField("count", len(result)).Debug("Multiple values retrieved from Redis")
	return result, nil
}

// Health проверяет состояние Redis
func (c *Client) Health(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	return err
}

// Загрузка в кеш активных заказов
func (c *Client) CacheWarmingOrders(db *database.DB) error {
	query := `SELECT * FROM orders WHERE status IN ($1, $2, $3, $4)
`
	rows, err := db.Query(query, models.OrderStatusAccepted, models.OrderStatusPreparing,
		models.OrderStatusReady, models.OrderStatusInDelivery)
	if err != nil {
		c.log.WithError(err).Error("Failed to make a SQL-query")
		return err
	}
	defer rows.Close()

	pipe := c.client.Pipeline()
	ctx := context.Background()
	for rows.Next() {
		order := &models.Order{}
		if err = rows.Scan(&order.ID, &order.CustomerName, &order.CustomerPhone,
			&order.PickupAddress, &order.DeliveryAddress, &order.TotalAmount, &order.DeliveryCost,
			&order.Status, &order.CourierID, &order.CreatedAt, &order.UpdatedAt, &order.DeliveredAt); err != nil {
			return fmt.Errorf("failed to scan orders: %w", err)
		}
		cacheKey := GenerateKey(KeyPrefixOrder, order.ID.String())
		data, err := json.Marshal(order)
		if err != nil {
			c.log.WithError(err).Error("Failed to marshal order")
			return err
		}
		pipe.Set(ctx, cacheKey, data, defaultCacheTTL)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	c.log.Info("Warmed up Redis cache with active orders")
	return nil
}

// Загрузка в кеш топ курьеров по рейтингу
func (c *Client) CacheWarmingCouriers(db *database.DB) error {
	query := `SELECT * FROM couriers WHERE rating >= $1 ORDER BY DESC LIMIT $2`
	rows, err := db.Query(query, minCourierRating, limitTopCouriers)
	if err != nil {
		c.log.WithError(err).Error("Failed to make a SQL-query")
		return err
	}
	defer rows.Close()

	pipe := c.client.Pipeline()
	ctx := context.Background()
	for rows.Next() {
		courier := &models.Courier{}
		if err = rows.Scan(&courier.ID, &courier.Name, &courier.Phone, &courier.Status,
			&courier.Rating, &courier.TotalReviews, &courier.CurrentLat, &courier.CurrentLon,
			&courier.CreatedAt, &courier.UpdatedAt, &courier.LastSeenAt); err != nil {
			return fmt.Errorf("failed to scan couriers: %w", err)
		}
		cacheKey := GenerateKey(KeyPrefixCourier, courier.ID.String())
		data, err := json.Marshal(courier)
		if err != nil {
			c.log.WithError(err).Error("Failed to marshal courier")
			return err
		}
		pipe.Set(ctx, cacheKey, data, defaultCacheTTL)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	c.log.Info("Warmed up Redis cache with top couriers")
	return nil
}

func (c *Client) Hit() { c.metrics.Hit() }

func (c *Client) Miss() { c.metrics.Miss() }

func (c *Client) GetMetrics(ctx context.Context) (uint64, uint64, int64, error) {
	hits := c.metrics.CacheMiss.Load()
	misses := c.metrics.CacheMiss.Load()
	cacheSize, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("Failed to load cache size: %w", err)
	}
	return hits, misses, cacheSize, err
}

// GenerateKey генерирует ключ для кеша
func GenerateKey(prefix, id string) string {
	return fmt.Sprintf("%s:%s", prefix, id)
}

// Константы для префиксов ключей
const (
	KeyPrefixOrder            = "order"
	KeyPrefixCourier          = "courier"
	KeyPrefixStats            = "stats"
	KeyPrefixOrderGeolocation = "order_geolocation"
	KeyPrefixReview           = "review"
)

// Константы, используемые при "прогреве" кеша
const (
	defaultCacheTTL  = 15 * time.Minute
	minCourierRating = 4
	limitTopCouriers = 50
)
