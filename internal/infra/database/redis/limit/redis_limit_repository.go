package limit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/entity/limit_entity"
	"github.com/redis/go-redis/v9"
)

type RedisLimitData struct {
	Id      string `redis:"id"`
	FreeAt  string `redis:"free_at"`
	LastAt  string `redis:"last_at"`
	Counter int32  `redis:"counter"`
}

type RedisLimitRepository struct {
	Rdb   *redis.Client
	Mutex *sync.Mutex
}

func NewRedisLimitRepository(host string, port string) *RedisLimitRepository {
	return &RedisLimitRepository{
		Rdb: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: "",
			DB:       0,
			Protocol: 2,
		}),
		Mutex: &sync.Mutex{},
	}
}

func (r *RedisLimitRepository) toRedis(limit *limit_entity.Limit) (*RedisLimitData, error) {
	var freeAtStr string
	if limit.FreeAt != nil {
		freeAtStr = limit.FreeAt.Format(time.RFC3339)
	}

	return &RedisLimitData{
		Id:      limit.Id,
		FreeAt:  freeAtStr,
		LastAt:  limit.LastAt.Format(time.RFC3339),
		Counter: limit.Counter,
	}, nil
}

func (r *RedisLimitRepository) toDomain(redisLimit *RedisLimitData) (*limit_entity.Limit, error) {
	var freeAt *time.Time
	var err error

	if redisLimit.FreeAt != "" {
		t, err := time.Parse(time.RFC3339, redisLimit.FreeAt)
		if err != nil {
			return &limit_entity.Limit{}, err
		}
		freeAt = &t
	}

	lastAt, err := time.Parse(time.RFC3339, redisLimit.LastAt)
	if err != nil {
		return &limit_entity.Limit{}, err
	}

	return &limit_entity.Limit{
		Id:      redisLimit.Id,
		FreeAt:  freeAt,
		LastAt:  lastAt,
		Counter: redisLimit.Counter,
	}, nil
}

func (r *RedisLimitRepository) CreateLimit(ctx context.Context, limit *limit_entity.Limit) error {
	redisData, err := r.toRedis(limit)
	if err != nil {
		println("REDIS: erro to redis")
		return err
	}

	_, err = r.Rdb.HSet(ctx, limit.Id, redisData).Result()

	if err != nil {
		println("REDIS: erro HSET")
		return err
	}

	return nil
}

func (r *RedisLimitRepository) GetLimitById(ctx context.Context, id string) (*limit_entity.Limit, error) {
	exists, err := r.Rdb.Exists(ctx, id).Result()
	if err != nil {
		return nil, err
	}

	if exists == 0 {
		return nil, nil
	}

	var redisData RedisLimitData
	err = r.Rdb.HGetAll(ctx, id).Scan(&redisData)

	if err != nil {
		println("REDIS: hgetall")
		println(err)
		return nil, err
	}

	limit, err := r.toDomain(&redisData)
	if err != nil {
		println("REDIS: toDomain")
		println(err)
		return nil, err
	}

	return limit, nil
}

func (r *RedisLimitRepository) UpdateLimitById(ctx context.Context, id string, limit *limit_entity.Limit) error {
	redisData, err := r.toRedis(limit)
	if err != nil {
		return err
	}

	_, err = r.Rdb.HSet(ctx, id, redisData).Result()

	if err != nil {
		return err
	}

	return nil
}
