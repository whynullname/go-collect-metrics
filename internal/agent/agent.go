// Пакет agent предоставляет основную логику агента.
package agent

import (
	"context"
	"encoding/json"
	"runtime"
	"sync"
	"time"

	"github.com/whynullname/go-collect-metrics/internal/agent/collector"
	"github.com/whynullname/go-collect-metrics/internal/agent/sender"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

type Agent struct {
	sender    *sender.AgentSender
	Collector *collector.AgentCollector
	config    *config.AgentConfig
}

func NewAgent(metricUseCase *metrics.MetricsUseCase, config *config.AgentConfig) *Agent {
	collector := collector.NewAgentCollector(&runtime.MemStats{}, metricUseCase)

	return &Agent{
		sender:    sender.NewAgentSender(collector, config),
		Collector: collector,
		config:    config,
	}
}

// UpdateMetrics горутина которая каждые config.PollInterval обновляет метрики в репозитории.
func (a *Agent) UpdateMetrics(ctx context.Context, wg *sync.WaitGroup) {
	updateDuration := time.Duration(a.config.PollInterval) * time.Second
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
			a.Collector.CollectMetrics()
		}
	}
}

// SendActualMetrics горутина которая каждые config.ReportInterval отправляет метрики на сервер.
func (a *Agent) SendActualMetrics(ctx context.Context, wg *sync.WaitGroup) {
	updateDuration := time.Duration(a.config.ReportInterval) * time.Second
	ticker := time.NewTicker(updateDuration)
	defer func() {
		ticker.Stop()
		wg.Done()
	}()

	var workerWaitGroup sync.WaitGroup
	jobs := make(chan *repository.Metric, 18)
	for i := 0; i < a.config.RateLimit; i++ {
		workerWaitGroup.Add(1)
		go a.worker(&workerWaitGroup, jobs)
	}
	for {
		select {
		case <-ctx.Done():
			close(jobs)
			workerWaitGroup.Wait()
			return
		case <-ticker.C:
			metricsArray, err := a.Collector.GetAllMetrics()
			if err != nil {
				continue
			}

			for _, json := range metricsArray {
				select {
				case jobs <- &json:
					// успешно отправлено
				default:
					logger.Log.Warnf("jobs channel is full, dropping metric %s\n", json.ID)
				}
			}
		}
	}
}

func (a *Agent) worker(wg *sync.WaitGroup, metricsToSend <-chan *repository.Metric) {
	defer wg.Done()

	for metric := range metricsToSend {
		jsonBytes, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Infof("error %s", err.Error())
		} else {
			logger.Log.Infof("Send metric %s\n", metric.ID)
			a.sender.SendJSONWithEncoding(jsonBytes, true)
		}
	}
}
