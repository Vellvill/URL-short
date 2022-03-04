package repository

import (
	"NewOne/internal/models"
	"NewOne/internal/postgres"
	"NewOne/internal/usecases"
	"NewOne/internal/utils"
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"log"
)

type repository struct {
	client postgres.Client
}

func NewRepository(client postgres.Client) usecases.Repository {
	return &repository{
		client: client,
	}
}

func (r *repository) AddLink(ctx context.Context, url *models.Url) error {
	if err := r.client.QueryRow(ctx, "insert into url (longurl, shorturl, numbersofredirect) values ($1, $2, $3) returning shorturl", url.Longurl, utils.Encode([]rune(url.Longurl)), url.Numbersofredirect).Scan(&url.Shorturl); err != nil {
		//может быть nil в контексте таблицы
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch {
			case pgErr.Code == "23505":
				r.client.QueryRow(ctx, "select shorturl from public.url where longurl = $1", url.Longurl).Scan(&url.Shorturl)
				return nil
			default:
				newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.SQLState()))
				return newErr
			}
		}
	}
	fmt.Println(url)
	return nil
}

func (r *repository) GetLink(ctx context.Context, shortUrl string) string {
	url := models.NewModelURL(0, "", shortUrl, 0)
	err := r.client.QueryRow(ctx, "SELECT longurl FROM public.url WHERE shorturl = $1", url.Shorturl).Scan(&url.Longurl)
	if err != nil {
		return ""
	}
	_, err = r.client.Exec(ctx, "update public.url set numbersofredirect = numbersofredirect + 1 where shorturl = $1", url.Shorturl)
	if err != nil {
		log.Fatal(err)
	}
	return url.Longurl
}

func (r *repository) GetStats(ctx context.Context, url *models.Url) error {
	err := r.client.QueryRow(ctx, "Select longurl, numbersofredirect from public.url where shorturl = $1", url.Shorturl).Scan(&url.Longurl, &url.Numbersofredirect)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.SQLState()))
			return newErr
		}
		return err
	}
	return nil
}
