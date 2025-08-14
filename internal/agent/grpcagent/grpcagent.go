package grpcagent

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/whynullname/go-collect-metrics/internal/agent/collector"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	pb "github.com/whynullname/go-collect-metrics/internal/proto"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCAgent struct {
	metricsClient pb.MetricsClient
	conn          *grpc.ClientConn
	collector     *collector.AgentCollector
	config        *config.AgentConfig
}

func NewGRPCAgent(metricUseCase *metrics.MetricsUseCase, config *config.AgentConfig) *GRPCAgent {
	collector := collector.NewAgentCollector(&runtime.MemStats{}, metricUseCase)

	return &GRPCAgent{
		collector: collector,
		config:    config,
	}
}

func (g *GRPCAgent) ConnetToServer() error {
	con, err := grpc.NewClient(g.config.EndPointAdress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	g.conn = con
	g.metricsClient = pb.NewMetricsClient(g.conn)
	return nil
}

func (g *GRPCAgent) UpdateMetrics(ctx context.Context, wg *sync.WaitGroup) {
	updateDuration := time.Duration(g.config.PollInterval) * time.Second
	ticker := time.NewTicker(updateDuration)
	defer func() {
		ticker.Stop()
		wg.Done()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Log.Info("Collect metrics")
			g.collector.CollectMetrics()
		}
	}
}

func (g *GRPCAgent) SendActualMetrics(ctx context.Context, wg *sync.WaitGroup) {
	updateDuration := time.Duration(g.config.ReportInterval) * time.Second
	ticker := time.NewTicker(updateDuration)
	defer func() {
		ticker.Stop()
		wg.Done()
	}()

	var workerWaitGroup sync.WaitGroup
	jobs := make(chan *repository.Metric, 18)
	for i := 0; i < g.config.RateLimit; i++ {
		workerWaitGroup.Add(1)
		go g.worker(&workerWaitGroup, jobs)
	}
	for {
		select {
		case <-ctx.Done():
			close(jobs)
			workerWaitGroup.Wait()
			return
		case <-ticker.C:
			metricsArray, err := g.collector.GetAllMetrics()
			if err != nil {
				continue
			}

			for _, metric := range metricsArray {
				jobs <- &metric
			}
		}
	}
}

func (g *GRPCAgent) worker(wg *sync.WaitGroup, metricsToSend <-chan *repository.Metric) {
	defer wg.Done()

	for metric := range metricsToSend {
		updateMetric := &pb.Metric{
			Id:    metric.ID,
			Type:  metric.MType,
			Delta: 0,
			Value: 0,
		}

		if metric.Delta != nil {
			updateMetric.Delta = *metric.Delta
		}

		if metric.Value != nil {
			updateMetric.Value = *metric.Value
		}

		resp, err := g.metricsClient.UpdateMetric(context.TODO(), &pb.UpdateMetricRequest{
			Metric: updateMetric,
		})

		if err != nil {
			logger.Log.Errorln(err)
			continue
		}

		if resp.Error != "" {
			logger.Log.Errorln(resp.Error)
		}
	}
}

func (g *GRPCAgent) CloseConnection() error {
	if g.conn == nil {
		return errors.New("grpc connection is nil")
	}

	err := g.conn.Close()
	return err
}
