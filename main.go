package main

import (
	"UrlShort/config"
	service "UrlShort/internal"
	middleware "UrlShort/internal/metrics"
	"UrlShort/internal/postgres"
	"UrlShort/internal/repository"
	"context"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func main() {

	cfg := config.GetConfig()

	client, err := postgres.NewClient(context.TODO(), cfg.Storage)
	if err != nil {
		log.Fatal("DB: Error with creating postgres client, err:", err)
	}

	repo, err := repository.NewRepository(client)
	if err == nil {
		log.Printf("REPO: New repository created")
	}

	impl, err := service.New(repo)
	if err == nil {
		log.Printf("IMPL: Implementation succsess")
	}

	r := mux.NewRouter()

	metricsMiddleware := middleware.NewMetricsMiddleware()

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/add", impl.AddNewUrl)

	r.HandleFunc("/{shorturl}", metricsMiddleware.Metrics(impl.RedirectToUrl))

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Println(err)
	}
}
