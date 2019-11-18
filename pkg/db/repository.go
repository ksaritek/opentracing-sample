package db

import (
	"context"
)

//Repository for db crud operation
type Repository interface {
	IsReady() error
	IsOk() error
	Upsert(context.Context, string, string) error
	Get(context.Context, string) (string, error)
}
