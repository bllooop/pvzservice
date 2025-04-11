package prometheus

import "github.com/prometheus/client_golang/prometheus"

var (
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
	prometheus.MustRegister(NumOfCreatedPVZ)
	prometheus.MustRegister(NumOfCreatedRecep)
	prometheus.MustRegister(NumOfAddedProducts)

}
