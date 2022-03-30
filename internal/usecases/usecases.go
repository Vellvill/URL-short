package usecases

import (
	"UrlShort/internal/models"
	"context"
)

type Repository interface {
	GetLink(ctx context.Context, shortUrl string) (string, error)
	AddLink(ctx context.Context, url *models.Url) error
	AddStartLink(ctx context.Context) error
}
