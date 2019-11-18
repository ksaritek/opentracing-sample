package db

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type db struct {
	client *redis.Client
}

func init() {
	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
}

//NewRepository is for redis connection
func NewRepository() Repository {

	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.address"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	return &db{client: client}
}

func (db *db) IsReady() error {
	if _, err := db.client.Ping().Result(); err != nil {
		return errors.Wrap(err, "could not connect to redis:")
	}

	return nil
}

func (db *db) IsOk() error {
	if _, err := db.client.Ping().Result(); err != nil {
		return errors.Wrap(err, "could not connect to redis:")
	}

	return nil
}

func (db *db) Upsert(ctx context.Context, key string, val string) error {
	tracer := opentracing.GlobalTracer()
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "redis-upsert")

	ext.SpanKindRPCClient.Set(span)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	if err := db.client.WithContext(ctx).Set(key, val, 0).Err(); err != nil {
		span.SetTag("error", true)
		span.LogFields(
			otlog.String("event", "upsert"),
			otlog.String("level", "error"),
			otlog.String("message", errors.Wrap(err, "upsert key at redis:").Error()),
		)
		return err
	}

	return nil
}

func (db *db) Get(ctx context.Context, key string) (string, error) {
	tracer := opentracing.GlobalTracer()
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "redis-get")

	ext.SpanKindRPCClient.Set(span)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	result, err := db.client.WithContext(ctx).Get(key).Result()

	if err != nil {
		span.SetTag("error", true)
		span.LogFields(
			otlog.String("event", "get"),
			otlog.String("level", "error"),
			otlog.String("message", errors.Wrap(err, "get key from redis:").Error()),
		)
		return "", err
	}
	return result, nil
}
