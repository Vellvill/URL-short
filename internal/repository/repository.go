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

func NewRepository(client postgres.Client) (usecases.Repository, error) {
	repo := &repository{
		client: client,
	}
	chStart := make(chan struct{}, 1)
	chDone := make(chan struct{})

	chStart <- struct{}{}

	log.Println("Starting status system")
	go func() {
		err := repo.Status(context.TODO(), chStart, chDone)
		if err != nil {
			log.Println(err)
		}
	}()
	startStatus(chStart, chDone)
	return repo, nil
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
	if err := r.client.QueryRow(ctx, "insert into url (longurl, shorturl) values ($1, $2) returning shorturl", url.Longurl, utils.Encode()).Scan(&url.Shorturl); err != nil {
		//может быть nil в контексте таблицы
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch {
			case pgErr.Code == "23505":
				err = r.client.QueryRow(ctx, "select shorturl from url where longurl = $1", url.Longurl).Scan(&url.Shorturl)
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
	url := models.NewModelURL(0, "", shortUrl, "")
	err := r.client.QueryRow(ctx, "SELECT longurl FROM url WHERE shorturl = $1", url.Shorturl).Scan(&url.Longurl)
	if err != nil {
		return "", err
	}
	return url.Longurl, nil
}

func (r *repository) GetStats(ctx context.Context, url *models.Url) error {
	err := r.client.QueryRow(ctx, "Select longurl, from url where shorturl = $1", url.Shorturl).Scan(&url.Longurl)
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
	rows, err := r.client.Query(ctx, "select id, longurl, shorturl, status from url")
	if err != nil {
		return nil, err
	}
	urls := make([]models.Url, 0)

	for rows.Next() {
		var url models.Url

		err = rows.Scan(&url.ID, &url.Longurl, &url.Shorturl, &url.Status)
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

func (r *repository) Status(ctx context.Context, ChStart <-chan struct{}, done chan<- struct{}) error {
	for {
		select {
		case <-ChStart:
			wg := new(sync.WaitGroup)
			rows, err := r.client.Query(ctx, "SELECT longurl FROM url order by url.id limit 300")
			if err != nil {
				return err
			}
			for rows.Next() {
				var url models.Url
				err = rows.Scan(&url.Longurl)
				if err != nil {
					return err
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
    						url 
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
