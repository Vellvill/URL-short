package main

import (
	"NewOne/config"
	service "NewOne/internal"
	middleware "NewOne/internal/metrics"
	"NewOne/internal/postgres"
	"NewOne/internal/repository"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
)

func main() {
	cfg := config.GetConfig()

	client, err := postgres.NewClient(context.TODO(), cfg.Storage)
	if err != nil {
		log.Fatal()
	}

	repo := repository.NewRepository(client)

	impl := service.New(repo)

	metricsMiddleware := middleware.NewMetricsMiddleware()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/add", impl.AddNewUrl)

	http.HandleFunc("/", metricsMiddleware.Metrics(impl.RedirectToUrl))

	http.HandleFunc("/check_status", impl.CheckStatus)

	start(cfg)
}

func start(cfg *config.Config) {
	listner, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIp, cfg.Listen.Port))
	if err != nil {
		log.Fatal()
	}

	if err = http.Serve(listner, nil); err != nil {
		log.Fatal()
	}

	log.Printf("server is listening %s:%s", cfg.Listen.BindIp, cfg.Listen.Port)
}
