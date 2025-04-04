package postgres

import (
	"context"
	"database/sql"

	"github.com/whynullname/go-collect-metrics/internal/logger"
)

type Postgres struct {
	DB     *sql.DB
	Adress string
}

const (
	GaugeMetricsTableName   = "gauge_metrics"
	CounterMetricsTableName = "counter_metrics"
)

func NewPostgresRepo(adress string) *Postgres {
	db, err := sql.Open("pgx", adress)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	_, err = db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS "+GaugeMetricsTableName+
		"(metric_id varchar(150) NOT NULL, metric_value double precision NOT NULL)")
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	_, err = db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS "+CounterMetricsTableName+
		"(metric_id varchar(150) NOT NULL, metric_value bigint NOT NULL)")
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	return &Postgres{
		DB:     db,
		Adress: adress,
	}
}

func (p *Postgres) GetGaugeMetricValue(metricName string) (float64, bool) {
	row := p.DB.QueryRowContext(context.Background(), "SELECT metric_value FROM "+GaugeMetricsTableName+" WHERE metric_id = $1", metricName)
	var output float64
	err := row.Scan(&output)
	if err != nil {
		logger.Log.Error(err)
	}

	return output, err != nil
}

func (p *Postgres) GetCounterMetricValue(metricName string) (int64, bool) {
	row := p.DB.QueryRowContext(context.Background(), "SELECT metric_value FROM "+CounterMetricsTableName+" WHERE metric_id = $1", metricName)
	var output int64
	err := row.Scan(&output)
	return output, err != nil
}

func (p *Postgres) UpdateGaugeMetricValue(metricName string, metricValue float64) float64 {
	_, err := p.DB.ExecContext(context.Background(), "INSERT INTO "+GaugeMetricsTableName+" (metric_id, metric_value) VALUES ($1, $2)", metricName, metricValue)
	if err != nil {
		logger.Log.Error(err)
		return 0
	}
	return metricValue
}

func (p *Postgres) UpdateCounterMetricValue(metricName string, metricValue int64) int64 {
	val, ok := p.GetCounterMetricValue(metricName)

	if !ok {
		val = metricValue
	} else {
		val += metricValue
	}

	_, err := p.DB.ExecContext(context.Background(), "INSERT INTO "+CounterMetricsTableName+" (metric_id, metric_value) VALUES ($1, $2)", metricName, metricValue)
	if err != nil {
		logger.Log.Error(err)
		return 0
	}
	return val
}

func (p *Postgres) GetAllGaugeMetrics() map[string]float64 {
	rows, err := p.DB.QueryContext(context.Background(), "SELECT * FROM "+GaugeMetricsTableName)
	if err != nil {
		logger.Log.Error(err)
		return map[string]float64{}
	}

	output := make(map[string]float64, 0)
	for rows.Next() {
		var metricName string
		var metricValue float64
		rows.Scan(&metricName, &metricValue)
		output[metricName] = metricValue
	}

	if rows.Err() != nil {
		logger.Log.Error(err)
		return map[string]float64{}
	}

	return output
}

func (p *Postgres) GetAllCounterMetrics() map[string]int64 {
	rows, err := p.DB.QueryContext(context.Background(), "SELECT * FROM "+CounterMetricsTableName)
	if err != nil {
		logger.Log.Error(err)
		return map[string]int64{}
	}

	output := make(map[string]int64, 0)
	for rows.Next() {
		var metricName string
		var metricValue int64
		rows.Scan(&metricName, &metricValue)
		output[metricName] = metricValue
	}

	if rows.Err() != nil {
		logger.Log.Error(err)
		return map[string]int64{}
	}

	return output
}

func (p *Postgres) CloseRepository() {
	p.DB.Close()
}
