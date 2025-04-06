package postgres

import (
	"context"
	"database/sql"

	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
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

func (p *Postgres) UpdateMetric(metric *repository.Metric) *repository.Metric {
	switch metric.MType {
	case repository.GaugeMetricKey:
		return p.UpdateGaugeMetric(metric)
	case repository.CounterMetricKey:
		return p.UpdateCounterMetricValue(metric)
	}
	return nil
}

func (p *Postgres) UpdateGaugeMetric(metric *repository.Metric) *repository.Metric {
	res, err := p.DB.ExecContext(context.Background(), "UPDATE "+GaugeMetricsTableName+` 
	SET metric_value = $1 WHERE metric_id = $2`, metric.Value, metric.ID)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		_, err = p.DB.ExecContext(context.Background(), "INSERT INTO "+GaugeMetricsTableName+` 
		(metric_id, metric_value) 
		VALUES ($1, $2)`, metric.ID, metric.Value)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	}

	return metric
}

func (p *Postgres) UpdateCounterMetricValue(metric *repository.Metric) *repository.Metric {
	val, ok := p.GetMetric(metric.ID, metric.MType)

	if !ok {
		val = metric
		_, err := p.DB.ExecContext(context.Background(), "INSERT INTO "+CounterMetricsTableName+" (metric_id, metric_value) VALUES ($1, $2)", val.ID, val.Delta)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	} else {
		newDelta := (*metric.Delta) + (*val.Delta)
		val.Delta = &newDelta
		_, err := p.DB.ExecContext(context.Background(), "UPDATE "+CounterMetricsTableName+" SET metric_value = $1 WHERE metric_id = $2", val.Delta, val.ID)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	}

	return val
}

func (p *Postgres) GetMetric(metricName string, metricType string) (*repository.Metric, bool) {
	tableName := p.GetTableNameByMetricType(metricType)
	row := p.DB.QueryRowContext(context.Background(), "SELECT metric_value FROM "+tableName+" WHERE metric_id = $1", metricName)
	output, err := p.ScanMetricByMetricType(row, metricType)
	return output, err == nil
}

func (p *Postgres) GetAllMetricsByType(metricType string) []repository.Metric {
	output := make([]repository.Metric, 0)
	tableName := p.GetTableNameByMetricType(metricType)
	rows, err := p.DB.QueryContext(context.Background(), "SELECT * FROM "+tableName)
	if err != nil {
		logger.Log.Error(err)
		return output
	}

	for rows.Next() {
		metric := repository.Metric{MType: metricType}
		switch metricType {
		case repository.GaugeMetricKey:
			err := rows.Scan(&metric.ID, &metric.Value)
			if err != nil {
				logger.Log.Error(err)
				return output
			}
			break
		case repository.CounterMetricKey:
			rows.Scan(&metric.ID, &metric.Delta)
			if err != nil {
				logger.Log.Error(err)
				return output
			}
			break
		}

		output = append(output, metric)
	}

	return output
}

func (p *Postgres) CloseRepository() {
	p.DB.Close()
}

func (p *Postgres) GetTableNameByMetricType(metricType string) string {
	switch metricType {
	case repository.CounterMetricKey:
		return CounterMetricsTableName
	case repository.GaugeMetricKey:
		return GaugeMetricsTableName
	default:
		return ""
	}
}

func (p *Postgres) ScanMetricByMetricType(row *sql.Row, metricType string) (output *repository.Metric, err error) {
	output = &repository.Metric{}
	switch metricType {
	case repository.CounterMetricKey:
		err = row.Scan(&output.Delta)
		if err != nil {
			logger.Log.Error(err)
			return
		}
	case repository.GaugeMetricKey:
		err = row.Scan(&output.Value)
		if err != nil {
			logger.Log.Error(err)
			return
		}
	}

	err = row.Err()
	if err != nil {
		logger.Log.Error(err)
	}
	return
}
