package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

// Cache defines the behavior for our cache implementations
type Cache interface {
	Set(key string, value string, expiration time.Duration) error
	Get(key string) (string, bool)
	Delete(key string) error
}

// instance holds the active cache implementation
var instance Cache

type Config struct {
	Type  CacheType   `json:"type" yaml:"type"`
	Redis ConfigRedis `json:"redis" yaml:"redis"`
}

type ConfigRedis struct {
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
}

type CacheType string

const (
	CacheMap   CacheType = "map"
	CacheRedis CacheType = "redis"
)

// Init initializes the global cache instance based on the configuration
func Init(config Config) {
	switch config.Type {
	case CacheMap:
		instance = newMemoryCache()
		log.Println("Initialized Memory Cache")
	case CacheRedis:
		instance = newRedisCache(config.Redis)
		log.Println("Initialized Redis Cache")
	default:
		instance = newMemoryCache()
		log.Println("Defaulting to Memory Cache")
	}
}

// SetCache stores a key-value pair using the initialized cache implementation.
func SetCache(key string, value string, expiration time.Duration) error {
	if instance == nil {
		return fmt.Errorf("cache not initialized: call cache.Init() first")
	}
	return instance.Set(key, value, expiration)
}

// UserWalletsKey returns the cache key for a user's wallet list.
func UserWalletsKey(userID int64) string { return fmt.Sprintf("wallets:%d", userID) }

// UserContactsKey returns the cache key for a user's contact list.
func UserContactsKey(userID int64) string { return fmt.Sprintf("contacts:%d", userID) }

// GetCache retrieves a value using the initialized cache implementation.
func GetCache(key string) (string, bool) {
	if instance == nil {
		return "", false
	}
	return instance.Get(key)
}

// DeleteCache removes a value using the initialized cache implementation.
func DeleteCache(key string) error {
	if instance == nil {
		return nil
	}
	return instance.Delete(key)
}

// ---------------------------------------------------------------------

type memoryCache struct {
	client *cache.Cache
}

func newMemoryCache() *memoryCache {
	return &memoryCache{
		client: cache.New(24*time.Hour, 24*time.Hour),
	}
}

func (m *memoryCache) Set(key string, value string, expiration time.Duration) error {
	m.client.Set(key, value, expiration)
	return nil
}

func (m *memoryCache) Get(key string) (string, bool) {
	val, found := m.client.Get(key)
	if !found {
		return "", false
	}
	return val.(string), true
}

func (m *memoryCache) Delete(key string) error {
	m.client.Delete(key)
	return nil
}

// ---------------------------------------------------------------------

type redisCache struct {
	client *redis.Client
}

func newRedisCache(config ConfigRedis) *redisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Username: config.User,
		Password: config.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return &redisCache{client: rdb}
}

func (r *redisCache) Set(key string, value string, expiration time.Duration) error {
	return r.client.Set(context.Background(), key, value, expiration).Err()
}

func (r *redisCache) Get(key string) (string, bool) {
	val, err := r.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", false
	} else if err != nil {
		logr.DefaultLogger.Errorw("redis cache", "err", err.Error())
		return "", false
	}
	return val, true
}

func (r *redisCache) Delete(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
