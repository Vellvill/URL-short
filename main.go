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
	"time"
)

func main() {

	for i := 0; i < 31; i++ {
		time.Sleep(1 * time.Second)
		log.Printf("Waiting %d/30", i)
	}

	cfg := config.GetConfig()

	client, err := postgres.NewClient(context.TODO(), cfg.Storage)
	if err != nil {
		log.Fatal("Error with creating postgres client, err:", err)
	} else {
		log.Printf("Succsess for connectiong to DB, storage cfg: %#v", cfg.Storage)
	}

	repo, err := repository.NewRepository(client)
	if err == nil {
		log.Printf("New repository created, %#v", repo)
	}

	impl, err := service.New(repo)
	if err == nil {
		log.Printf("Implementation succsess")
	}

	metricsMiddleware := middleware.NewMetricsMiddleware()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/add", impl.AddNewUrl)

	http.HandleFunc("/", metricsMiddleware.Metrics(impl.RedirectToUrl))

	http.HandleFunc("/check_status", impl.CheckStatus)

	err = start(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server is listening %s:%s", cfg.Listen.BindIp, cfg.Listen.Port)
}

func start(cfg *config.Config) error {
	listner, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIp, cfg.Listen.Port))
	if err != nil {
		return err
	}

	if err = http.Serve(listner, nil); err != nil {
		return err
	}
	return nil
}
