package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/mem"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

type Agent struct {
	memStats       *runtime.MemStats
	Config         *config.AgentConfig
	Client         *resty.Client
	metricsUseCase *metrics.MetricsUseCase
}

func NewAgent(memStats *runtime.MemStats, metricUseCase *metrics.MetricsUseCase, config *config.AgentConfig) *Agent {
	client := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return err != nil || r.StatusCode() >= http.StatusInternalServerError
			},
		)
	return &Agent{
		memStats:       memStats,
		metricsUseCase: metricUseCase,
		Config:         config,
		Client:         client,
	}
}

func (a *Agent) worker(metricsToSend <-chan *repository.Metric) {
	for metric := range metricsToSend {
		jsonBytes, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Infof("error %s", err.Error())
			continue
		}
		logger.Log.Infof("Send metric %s\n", metric.ID)
		a.sendJSONWithEncoding(jsonBytes)
	}
}

func (a *Agent) UpdateMetrics() {
	updateDuration := time.Duration(a.Config.PollInterval) * time.Second
	for {
		ticker := time.NewTicker(updateDuration)
		<-ticker.C
		ticker.Stop()

		logger.Log.Info("Update metrics")
		memStats := a.memStats
		runtime.ReadMemStats(memStats)
		a.UpdateGaugeMetricValue("Alloc", float64(memStats.Alloc))
		a.UpdateGaugeMetricValue("Frees", float64(memStats.Frees))
		a.UpdateGaugeMetricValue("BuckHashSys", float64(memStats.BuckHashSys))
		a.UpdateGaugeMetricValue("GCCPUFraction", float64(memStats.GCCPUFraction))
		a.UpdateGaugeMetricValue("GCSys", float64(memStats.GCSys))
		a.UpdateGaugeMetricValue("HeapAlloc", float64(memStats.HeapAlloc))
		a.UpdateGaugeMetricValue("HeapIdle", float64(memStats.HeapIdle))
		a.UpdateGaugeMetricValue("HeapInuse", float64(memStats.HeapInuse))
		a.UpdateGaugeMetricValue("HeapObjects", float64(memStats.HeapObjects))
		a.UpdateGaugeMetricValue("HeapReleased", float64(memStats.HeapReleased))
		a.UpdateGaugeMetricValue("HeapSys", float64(memStats.HeapSys))
		a.UpdateGaugeMetricValue("LastGC", float64(memStats.LastGC))
		a.UpdateGaugeMetricValue("Lookups", float64(memStats.Lookups))
		a.UpdateGaugeMetricValue("MCacheSys", float64(memStats.MCacheSys))
		a.UpdateGaugeMetricValue("Mallocs", float64(memStats.Mallocs))
		a.UpdateGaugeMetricValue("NextGC", float64(memStats.NextGC))
		a.UpdateGaugeMetricValue("NumForcedGC", float64(memStats.NumForcedGC))
		a.UpdateGaugeMetricValue("NumGC", float64(memStats.NumGC))
		a.UpdateGaugeMetricValue("OtherSys", float64(memStats.OtherSys))
		a.UpdateGaugeMetricValue("PauseTotalNs", float64(memStats.PauseTotalNs))
		a.UpdateGaugeMetricValue("StackInuse", float64(memStats.StackInuse))
		a.UpdateGaugeMetricValue("StackSys", float64(memStats.StackSys))
		a.UpdateGaugeMetricValue("Sys", float64(memStats.Sys))
		a.UpdateGaugeMetricValue("TotalAlloc", float64(memStats.TotalAlloc))
		a.UpdateGaugeMetricValue("MCacheInuse", float64(memStats.MCacheInuse))
		a.UpdateGaugeMetricValue("MSpanInuse", float64(memStats.MSpanInuse))
		a.UpdateGaugeMetricValue("MSpanSys", float64(memStats.MSpanSys))
		a.UpdateGaugeMetricValue("RandomValue", rand.Float64())
		a.UpdateCounterMetricValue("PollCount", int64(1))

		v, err := mem.VirtualMemory()
		if err != nil {
			logger.Log.Error(err)
			continue
		}

		a.UpdateGaugeMetricValue("TotalMemory", float64(v.Total))
		a.UpdateGaugeMetricValue("FreeMemory", float64(v.Free))
		a.UpdateGaugeMetricValue("CPUutilization1", v.UsedPercent)
	}
}

func (a *Agent) SendActualMetrics() {
	updateDuration := time.Duration(a.Config.ReportInterval) * time.Second
	jobs := make(chan *repository.Metric, 18)
	for i := 0; i < a.Config.RateLimit; i++ {
		go a.worker(jobs)
	}
	for {
		ticker := time.NewTicker(updateDuration)
		<-ticker.C
		ticker.Stop()

		gaugeMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.GaugeMetricKey)
		if err != nil {
			return
		}
		counterMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.CounterMetricKey)
		if err != nil {
			return
		}

		metricsArray := append(gaugeMetrics, counterMetrics...)
		for _, json := range metricsArray {
			jobs <- &json
		}
	}
}

func (a *Agent) UpdateGaugeMetricValue(metricID string, value float64) {
	metric := repository.Metric{
		MType: repository.GaugeMetricKey,
		Value: &value,
		ID:    metricID,
	}
	a.metricsUseCase.UpdateMetric(context.TODO(), &metric)
}

func (a *Agent) UpdateCounterMetricValue(metricID string, value int64) {
	metric := repository.Metric{
		MType: repository.CounterMetricKey,
		Delta: &value,
		ID:    metricID,
	}
	a.metricsUseCase.UpdateMetric(context.TODO(), &metric)
}

func (a *Agent) SendMetrics() {
	gaugeMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.GaugeMetricKey)
	if err != nil {
		return
	}
	counterMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.CounterMetricKey)
	if err != nil {
		return
	}
	jsonArray := append(counterMetrics, gaugeMetrics...)
	a.sendPostResponseWithMetrics(jsonArray)
}

func (a *Agent) SendAllMetricsByArray() {
	gaugeMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.GaugeMetricKey)
	if err != nil {
		return
	}
	counterMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.CounterMetricKey)
	if err != nil {
		return
	}

	jsonArray := append(gaugeMetrics, counterMetrics...)
	url := fmt.Sprintf("http://%s/updates", a.Config.EndPointAdress)
	newRequest := a.Client.R().SetBody(jsonArray)
	_, err = newRequest.Post(url)
	if err != nil {
		logger.Log.Infof("error %s", err.Error())
	}
}

func (a *Agent) SendAllMetricByArrayAndSHA() {
	gaugeMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.GaugeMetricKey)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	counterMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.CounterMetricKey)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	jsonArray := append(gaugeMetrics, counterMetrics...)
	url := fmt.Sprintf("http://%s/updates", a.Config.EndPointAdress)
	hash := hmac.New(sha256.New, []byte(a.Config.HashKey))
	jsonBytes, err := json.Marshal(jsonArray)
	if err != nil {
		logger.Log.Infof("error %s", err.Error())
	}
	hash.Write(jsonBytes)
	requestHash := hex.EncodeToString(hash.Sum(nil))
	newRequest := a.Client.R().SetBody(jsonArray).
		SetHeader("HashSHA256", string(requestHash))
	_, err = newRequest.Post(url)
	if err != nil {
		logger.Log.Infof("error %s", err.Error())
	}
}

func (a *Agent) SendMetricsByJSON() {
	gaugeMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.GaugeMetricKey)
	if err != nil {
		return
	}
	counterMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.CounterMetricKey)
	if err != nil {
		return
	}

	jsonArray := append(gaugeMetrics, counterMetrics...)
	for _, metric := range jsonArray {
		jsonBytes, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Infof("error %s", err.Error())
			continue
		}
		a.sendJSONWithEncoding(jsonBytes)
	}
}

func (a *Agent) sendJSONWithEncoding(json []byte) {
	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	gz.Write(json)
	gz.Close()
	url := fmt.Sprintf("http://%s/update", a.Config.EndPointAdress)
	newRequest := a.Client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(&buff)

	_, err := newRequest.Post(url)
	if err != nil {
		logger.Log.Infof("error %s", err.Error())
	}
}

func (a *Agent) sendPostResponseWithMetrics(metrics []repository.Metric) {
	for _, metric := range metrics {
		metricValue := ""

		switch metric.MType {
		case repository.CounterMetricKey:
			metricValue = strconv.FormatInt(*metric.Delta, 10)
		case repository.GaugeMetricKey:
			metricValue = strconv.FormatFloat(*metric.Value, 'f', 2, 64)
		}

		url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.Config.EndPointAdress, metric.MType, metric.ID, metricValue)
		requst := a.Client.NewRequest()
		requst.SetHeader("ContentType", "text/plain")
		_, err := requst.Post(url)
		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}
	}
}
