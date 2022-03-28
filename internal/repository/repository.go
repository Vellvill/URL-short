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
	"sync"
	"time"
)

type repository struct {
	client postgres.Client
}

func NewRepository(client postgres.Client) usecases.Repository {
	repo := &repository{
		client: client,
	}
	chStart := make(chan struct{}, 1)
	chDone := make(chan struct{})

	chStart <- struct{}{}
	go repo.Status(context.TODO(), chStart, chDone)
	startStatus(chStart, chDone)
	return repo
}

func startStatus(chStart chan<- struct{}, chDone <-chan struct{}) {
	go func() {
		for {
			select {
			case <-chDone:
				chStart <- struct{}{}
			default:
				time.Sleep(10 * time.Minute)
			}
		}
	}()
}

func (r *repository) AddLink(ctx context.Context, url *models.Url) error {
	if err := r.client.QueryRow(ctx, "insert into url (longurl, shorturl, numbersofredirect) values ($1, $2, $3) returning shorturl", url.Longurl, utils.Encode(), url.Numbersofredirect).Scan(&url.Shorturl); err != nil {
		//может быть nil в контексте таблицы
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch {
			case pgErr.Code == "23505":
				err = r.client.QueryRow(ctx, "select shorturl from public.url where longurl = $1", url.Longurl).Scan(&url.Shorturl)
				if err != nil {
					log.Println(err)
				}
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

func (r *repository) GetLink(ctx context.Context, shortUrl string) (string, error) {
	url := models.NewModelURL(0, "", shortUrl, 0, "")
	err := r.client.QueryRow(ctx, "SELECT longurl FROM public.url WHERE shorturl = $1", url.Shorturl).Scan(&url.Longurl)
	if err != nil {
		return "", err
	}
	_, err = r.client.Exec(ctx, "update public.url set numbersofredirect = numbersofredirect + 1 where shorturl = $1", url.Shorturl)
	if err != nil {
		log.Println(err)
	}
	return url.Longurl, nil
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

func (r *repository) FindAll(ctx context.Context) (u []models.Url, err error) {
	rows, err := r.client.Query(ctx, "select id, longurl, shorturl, numbersofredirect, status from public.url")
	if err != nil {
		return nil, err
	}
	urls := make([]models.Url, 0)

	for rows.Next() {
		var url models.Url

		err = rows.Scan(&url.ID, &url.Longurl, &url.Shorturl, &url.Numbersofredirect, &url.Status)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *repository) Status(ctx context.Context, ChStart <-chan struct{}, done chan<- struct{}) {
	for {
		select {
		case <-ChStart:
			wg := new(sync.WaitGroup)
			rows, err := r.client.Query(ctx, "SELECT longurl FROM public.url order by url.id limit 300")
			if err != nil {
				log.Println(err)
			}
			for rows.Next() {
				var url models.Url
				err = rows.Scan(&url.Longurl)
				if err != nil {
					panic(err)
				}
				go func(url models.Url) {
					defer wg.Done()
					wg.Add(1)
					err = utils.Check(&url)
					if err != nil {
						log.Println(err)
					}
					q := `
						UPDATE 
    						public.url 
						SET 
    						status = $1 
						where 
      						longurl = $2
`
					if _, err = r.client.Exec(ctx, q, url.Status, url.Longurl); err != nil {
						log.Println(err)
					}
				}(url)
			}
			wg.Wait()
			done <- struct{}{}
		default:
			time.Sleep(10 * time.Minute)
		}
	}
}
