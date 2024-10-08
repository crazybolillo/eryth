package metric

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

var dbConnAcquired = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "dbpool_acquired_connections",
	Help: "Number of connections currently acquired.",
})
var dbConnIdle = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "dbpool_idle_connections",
	Help: "Number of idle connections currently idle.",
})
var dbConnConstructing = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "dbpool_constructing_connections",
	Help: "Number of connections being constructed.",
})
var dbMaxConns = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "dbpool_max_connections",
	Help: "Maximum amount of connections permitted.",
})

func updateDbPool(pool *pgxpool.Pool) {
	stat := pool.Stat()
	dbConnAcquired.Set(float64(stat.AcquiredConns()))
	dbConnIdle.Set(float64(stat.IdleConns()))
	dbConnConstructing.Set(float64(stat.ConstructingConns()))
	dbMaxConns.Set(float64(stat.MaxConns()))
}

func WatchDbPool(ctx context.Context, pool *pgxpool.Pool, refresh time.Duration) {
	updateDbPool(pool)
	ticker := time.NewTicker(refresh)
	go func() {
		for {
			select {
			case <-ticker.C:
				updateDbPool(pool)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
