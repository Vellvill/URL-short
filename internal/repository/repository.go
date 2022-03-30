package repository

import (
	"UrlShort/internal/models"
	"UrlShort/internal/postgres"
	"UrlShort/internal/usecases"
	"UrlShort/internal/utils"
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"log"
	"sync"
	"time"
)

type repository struct {
	sync.Mutex
	client postgres.Client
}

func NewRepository(client postgres.Client) (usecases.Repository, error) {
	repo := &repository{
		client: client,
	}
	err := repo.AddStartLink(context.Background())
	if err != nil {
		log.Println(err)
	}
	chStart := make(chan struct{}, 1)
	chDone := make(chan struct{})

	chStart <- struct{}{}

	log.Println("STATUS SYSTEM: Starting status system")
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

func (r *repository) AddStartLink(ctx context.Context) error {
	var added string
	q := `
	INSERT INTO url
		(longurl, shorturl)
	VALUES
		($1, $2)
	ON CONFLICT DO NOTHING
	RETURNING longurl
	`
	if err := r.client.QueryRow(ctx, q, "https://yandex.ru/", "11111").Scan(&added); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.SQLState()))
			return newErr
		} else {
			return err
		}
	}
	log.Println("Added 1 start link, link:%s", added)
	return nil
}

func (r *repository) AddLink(ctx context.Context, url *models.Url) error {
	q := `
	insert into url 
		(longurl, shorturl) 
	values 
		($1, $2) 
	returning shorturl
`
	if err := r.client.QueryRow(ctx, q, url.Longurl, utils.Encode()).Scan(&url.Shorturl); err != nil {
		//может быть nil в контексте таблицы
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch {
			case pgErr.Code == "23505":
				qErr := `
	select 
		shorturl 
	from url 
		where longurl = $1`
				err = r.client.QueryRow(ctx, qErr, url.Longurl).Scan(&url.Shorturl)
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
					r.Lock()
					if _, err = r.client.Exec(ctx, q, url.Status, url.Longurl); err != nil {
						log.Println(err)
					}
					r.Unlock()
				}(url)
			}
			wg.Wait()
			done <- struct{}{}
		default:
			time.Sleep(10 * time.Minute)
		}
	}
}
