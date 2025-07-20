// Пакет sender предназначен для отправки метрик на сервер.
package sender

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/whynullname/go-collect-metrics/internal/agent/collector"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type AgentSender struct {
	collector *collector.AgentCollector
	client    *resty.Client
	config    *config.AgentConfig
}

func NewAgentSender(collector *collector.AgentCollector, config *config.AgentConfig) *AgentSender {
	client := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return err != nil || r.StatusCode() >= http.StatusInternalServerError
			},
		)
	return &AgentSender{
		collector: collector,
		config:    config,
		client:    client,
	}
}

// SendAllMetricsByArray отправить все метрики одним массивом.
func (s *AgentSender) SendAllMetricsByArray() {
	jsonArray, err := s.collector.GetAllMetrics()
	if err != nil {
		return
	}

	url := fmt.Sprintf("http://%s/updates", s.config.EndPointAdress)
	newRequest := s.client.R().SetBody(jsonArray)
	s.sendRequest(newRequest, url)
}

// SendAllMetricByArrayAndSHA отправить все метрики массивом и подписать с помощью SHA.
func (s *AgentSender) SendAllMetricByArrayAndSHA() {
	jsonArray, err := s.collector.GetAllMetrics()
	if err != nil {
		return
	}

	url := fmt.Sprintf("http://%s/updates", s.config.EndPointAdress)
	jsonBytes, err := json.Marshal(jsonArray)
	if err != nil {
		logger.Log.Infof("error %s", err.Error())
	}
	requestHash := s.generateHash(jsonBytes)
	newRequest := s.client.R().SetBody(jsonArray).
		SetHeader("HashSHA256", requestHash)
	s.sendRequest(newRequest, url)
}

func (s *AgentSender) generateHash(data []byte) string {
	hash := hmac.New(sha256.New, []byte(s.config.HashKey))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// SendMetricsByJSON отправить все метрики в формате JSON.
// ВАЖНО! Каждая метрика отправляется по очереди.
// Метод отправляет json закодированным с помощью gzip.
func (s *AgentSender) SendMetricsByJSON() {
	jsonArray, err := s.collector.GetAllMetrics()
	if err != nil {
		return
	}

	for _, metric := range jsonArray {
		jsonBytes, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Infof("error %s", err.Error())
			continue
		}
		s.SendJSONWithEncoding(jsonBytes, true)
	}
}

// SendJSONWithEncoding позволяет отправить JSON в закодированном ввиде с помощью gzip.
func (s *AgentSender) SendJSONWithEncoding(json []byte, enableEncoding bool) {
	buff := s.GZIPData(json)
	url := fmt.Sprintf("http://%s/update", s.config.EndPointAdress)
	newRequest := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buff)

	s.sendRequest(newRequest, url)
}

// GZIPData кодирует массив byte с помощью gzip.
func (s *AgentSender) GZIPData(data []byte) *bytes.Buffer {
	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	gz.Write(data)
	gz.Close()
	return &buff
}

func (s *AgentSender) EncryptData(data []byte) []byte {
	if s.config.RSAKey == nil {
		logger.Log.Warnf("Can't encrypt data, because rsa key is nil")
		return data
	}

	encryptedMessage, err := rsa.EncryptPKCS1v15(rand.Reader, s.config.RSAKey, data)
	if err != nil {
		logger.Log.Errorf("Error while encrypt data: %v\n", err)
		return data
	}

	return encryptedMessage
}

// SendMetricsByPostResponse отправить каждую метрику с помощью POST формата по URL.
func (s *AgentSender) SendMetricsByPostResponse() {
	jsonArray, err := s.collector.GetAllMetrics()
	if err != nil {
		return
	}
	for _, metric := range jsonArray {
		metricValue := ""

		switch metric.MType {
		case repository.CounterMetricKey:
			metricValue = strconv.FormatInt(*metric.Delta, 10)
		case repository.GaugeMetricKey:
			metricValue = strconv.FormatFloat(*metric.Value, 'f', 2, 64)
		}

		url := fmt.Sprintf("http://%s/update/%s/%s/%s", s.config.EndPointAdress, metric.MType, metric.ID, metricValue)
		requst := s.client.NewRequest()
		requst.SetHeader("ContentType", "text/plain")
		_, err := requst.Post(url)
		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}
	}
}

func (s *AgentSender) sendRequest(request *resty.Request, url string) {
	_, err := request.Post(url)
	if err != nil {
		logger.Log.Infof("error %s", err.Error())
	}
}
