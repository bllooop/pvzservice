package prometheus

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP запросов",
		},
		[]string{"method", "path", "status"},
	)
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Продолжительность HTTP запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	NumOfCreatedPVZ = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_pvz_amount_total",
			Help: "Суммарное количество созданных ПВЗ",
		},
	)
	NumOfCreatedRecep = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_receptions_amount_total",
			Help: "Суммарное количество созданных приемок",
		},
	)
	NumOfAddedProducts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "added_products_amount_total",
			Help: "Суммарное количество добавленных товаров",
		},
	)
)

func init() {
	prometheus.MustRegister(HTTPRequestTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(NumOfCreatedPVZ)
	prometheus.MustRegister(NumOfCreatedRecep)
	prometheus.MustRegister(NumOfAddedProducts)
}
