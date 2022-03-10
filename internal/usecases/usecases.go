package usecases

import (
	"NewOne/internal/models"
	"context"
)

type Repository interface {
	GetLink(ctx context.Context, shortUrl string) (string, error)
	AddLink(ctx context.Context, url *models.Url) error
	GetStats(ctx context.Context, url *models.Url) error
	FindAll(ctx context.Context) (u []models.Url, err error)
}
