package metrics

import (
	"context"
	"log"
	"testing"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func BenchmarkMetricsUseCase(b *testing.B) {
	cfg := config.NewServerConfig()
	cfg.ParseFlags()
	repo, err := postgres.NewPostgresRepo(cfg.PostgressAdress)
	if err != nil {
		log.Fatal(err)
		return
	}

	useCase := NewMetricUseCase(repo)
	testDelta := int64(41)
	metricJSON := &repository.Metric{
		ID:    "test",
		MType: "counter",
		Delta: &testDelta,
	}

	b.Run("Update metric", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			useCase.UpdateMetric(context.Background(), metricJSON)
		}
	})

	b.Run("Update metrics", func(b *testing.B) {
		metrics := make([]repository.Metric, 100)
		for i := 0; i < 100; i++ {
			metrics = append(metrics, *metricJSON)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			useCase.UpdateMetrics(context.Background(), metrics)
		}
	})

	b.Run("Get metric", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			useCase.GetMetric(context.Background(), metricJSON.MType, metricJSON.ID)
		}
	})

	b.Run("Get all metrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			useCase.GetAllMetricsByType(context.Background(), metricJSON.MType)
		}
	})
}
