# url-short
This is my link shortener server using mux framework, postgres database, and prometheus + grafana bundle to display metrics

## Handlers
```golang
func (i *Implementation) RedirectToUrl(w http.ResponseWriter, r *http.Request)
```
You can access this handler by using `localhost:8080/{shorturl}` and you will be provided to your short url address

```golang
func (i *Implementation) AddNewUrl(w http.ResponseWriter, r *http.Request)
```
This handler allow you to add new link to postgres database, you can use it by routing `localhost:8080/add?url={your_long_url}`

## Migrations
When you starting up server for the first time it will create a new database with this scheme:
```sql
CREATE TABLE IF NOT EXISTS url(
    id SERIAL PRIMARY KEY not null,
    longurl VARCHAR(255) not null,
    shorturl VARCHAR(255),
    status VARCHAR(255)
);
```
Next function will create first start link which is `http://yandex.ru`
```golang
func (r *repository) AddStartLink(ctx context.Context) error 
```
## Status system
This is a gorutine system which checking status of url's every 10 minutes. It starts by sending an empty struct to special channel.
```golang
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
```
The status system takes 300 links, which are ordered by url.id.
## Prometheus and grafana
Prometheus is scraping every 10 minutes (btw you can change it in `prometheus -> prometheus.yml` which is config) every handlers using, and also it scrap count of redirects. In docker-compose there is also grafana for visualization metrics, but you can use `localhost:8080/metrics` for checking it out.
![image](https://user-images.githubusercontent.com/93131551/160903555-27215209-2928-48d6-b3e8-422e8d6e7689.png)

metrics:
![image](https://user-images.githubusercontent.com/93131551/160904044-76e18443-97e3-4a0a-b6de-205e995ba889.png)

## pgadmin
You can also use pgadmin for administrating your db. It starts on 5050 port.
## docker-compose
docker-compose starts a few images (postgres, server, pgadmin, grafana and prometheus), all dependenses you can find at docker-compose file. Keep on mind that while docker creating new images and volume for db it may take a while and this can take up a lot of space on your hard drive.
## Connecting to postgres
I used to connect to db by pgx.pool, so it creates a pool of connections.
```golang
func NewClient(ctx context.Context, cfg config.Storage) (pool *pgxpool.Pool, err error)
```
this function makes `5 attempts` to create a pool `every 5 seconds`. There is a problem that db could start after server, but if function will be out of attempts it will shut down because of panic using `nil pointer` in querys, but it will restart, because of `docker-compose` `restart: on-failure`.
## Starting up
- Clone repository
- run `docker-compose up -d` in terminal.

