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
	ch := make(chan struct{})
	repo.Status(context.TODO(), ch)
	startStatus(ch)
	return repo

}

func startStatus(startSt chan struct{}) {
	startSt <- struct{}{}
	for {
		select {
		case <-time.After(10 * time.Minute):
			startSt <- struct{}{}
		}
	}
}

func (r *repository) AddLink(ctx context.Context, url *models.Url) error {
	if err := r.client.QueryRow(ctx, "insert into url (longurl, shorturl, numbersofredirect, status) values ($1, $2, $3, $4) returning shorturl", url.Longurl, utils.Encode(), url.Numbersofredirect, url.Status).Scan(&url.Shorturl); err != nil {
		//может быть nil в контексте таблицы
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch {
			case pgErr.Code == "23505":
				err = r.client.QueryRow(ctx, "select shorturl from public.url where longurl = $1", url.Longurl).Scan(&url.Shorturl)
				if err != nil {
					log.Fatal(err)
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
		log.Fatal(err)
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

func (r *repository) Status(ctx context.Context, start <-chan struct{}) {
	for {
		select {
		case <-start:
			urlchan := make(chan models.Url)
			answer := make(chan models.Url)
			mu := new(sync.Mutex)
			wg := new(sync.WaitGroup)
			go func(wg *sync.WaitGroup, mu *sync.Mutex) {
				rows, err := r.client.Query(ctx, "SELECT longurl FROM public.url order by url.id limit 300")
				if err != nil {
					log.Fatal(err)
				}
				for rows.Next() {
					wg.Add(1)
					var url models.Url
					mu.Lock()
					err = rows.Scan(&url.Longurl)
					urlchan <- url
					mu.Unlock()
					wg.Done()
					if err != nil {
						log.Fatal(err)
					}
				}
			}(wg, mu)
			go func(answer chan models.Url, urlchan chan models.Url, wg *sync.WaitGroup) {
				wg.Wait()
				for {
					select {
					case x := <-urlchan:
						err := utils.Check(&x)
						if err != nil {
							fmt.Println(err)
						}
						answer <- x
					default:
						time.Sleep(1 * time.Second)
					}
				}
			}(answer, urlchan, wg)
			go func(answer chan models.Url, wg *sync.WaitGroup) {
				wg.Wait()
				for {
					select {
					case x := <-answer:
						_, err := r.client.Exec(ctx, "update public.url set status = $1 where longurl = $2", x.Status, x.Longurl)
						if err != nil {
							log.Fatal(err)
						}
					default:
						time.Sleep(1 * time.Second)
					}
				}
			}(answer, wg)
		default:
			time.Sleep(1 * time.Minute)
		}
	}
}
