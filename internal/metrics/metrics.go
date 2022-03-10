package metrics

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	redirects prometheus.GaugeVec
}

//Метрика, которая считает кол-во переходов по шорт-урлу, в идеале сделать миддлвейр над ручкой по переходу по шорт урлу 	http.HandleFunc("/", impl.RedirectToUrl)
