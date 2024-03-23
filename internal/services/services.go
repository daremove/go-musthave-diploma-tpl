package services

import "context"

type Storage interface {
	BeginTx(ctx context.Context)
}
