package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type Postgres struct {
	db                      *sql.DB
	selectGaugeMetricStmt   *sql.Stmt
	selectCounterMetricStmt *sql.Stmt
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

	selectGaugeStmt, err := db.PrepareContext(context.Background(), "SELECT metric_value FROM "+GaugeMetricsTableName+" WHERE metric_id = $1")
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	selectCounterStmt, err := db.PrepareContext(context.Background(), "SELECT metric_value FROM "+CounterMetricsTableName+" WHERE metric_id = $1")
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	instance := Postgres{
		db:                      db,
		selectGaugeMetricStmt:   selectGaugeStmt,
		selectCounterMetricStmt: selectCounterStmt,
	}
	return &instance
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

func (p *Postgres) UpdateMetrics(metrics []repository.Metric) ([]repository.Metric, error) {
	output := make([]repository.Metric, 0)
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}

	for _, metric := range metrics {
		var ouputMetric *repository.Metric
		switch metric.MType {
		case repository.GaugeMetricKey:
			ouputMetric = p.UpdateGaugeMetricWithTx(tx, &metric)
		case repository.CounterMetricKey:
			ouputMetric = p.UpdateCounterMetricValueWithTx(tx, &metric)
		}

		if ouputMetric == nil {
			tx.Rollback()
			return nil, err
		} else {
			output = append(output, *ouputMetric)
		}
	}

	tx.Commit()
	return output, nil
}

func (p *Postgres) UpdateGaugeMetric(metric *repository.Metric) *repository.Metric {
	res, err := p.db.ExecContext(context.Background(), "UPDATE "+GaugeMetricsTableName+` 
	SET metric_value = $1 WHERE metric_id = $2`, metric.Value, metric.ID)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		_, err = p.db.ExecContext(context.Background(), "INSERT INTO "+GaugeMetricsTableName+` 
		(metric_id, metric_value) 
		VALUES ($1, $2)`, metric.ID, metric.Value)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	}

	return metric
}

func (p *Postgres) UpdateGaugeMetricWithTx(tx *sql.Tx, metric *repository.Metric) *repository.Metric {
	res, err := tx.ExecContext(context.Background(), "UPDATE "+GaugeMetricsTableName+` 
	SET metric_value = $1 WHERE metric_id = $2`, metric.Value, metric.ID)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		_, err = tx.ExecContext(context.Background(), "INSERT INTO "+GaugeMetricsTableName+` 
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
		_, err := p.db.ExecContext(context.Background(), "INSERT INTO "+CounterMetricsTableName+" (metric_id, metric_value) VALUES ($1, $2)", val.ID, val.Delta)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	} else {
		newDelta := (*metric.Delta) + (*val.Delta)
		val.Delta = &newDelta
		_, err := p.db.ExecContext(context.Background(), "UPDATE "+CounterMetricsTableName+" SET metric_value = $1 WHERE metric_id = $2", val.Delta, val.ID)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	}

	return val
}

func (p *Postgres) UpdateCounterMetricValueWithTx(tx *sql.Tx, metric *repository.Metric) *repository.Metric {
	val, ok := p.GetMetric(metric.ID, metric.MType)

	if !ok {
		val = metric
		_, err := tx.ExecContext(context.Background(), "INSERT INTO "+CounterMetricsTableName+" (metric_id, metric_value) VALUES ($1, $2)", val.ID, val.Delta)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	} else {
		newDelta := (*metric.Delta) + (*val.Delta)
		val.Delta = &newDelta
		_, err := tx.ExecContext(context.Background(), "UPDATE "+CounterMetricsTableName+" SET metric_value = $1 WHERE metric_id = $2", val.Delta, val.ID)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	}

	return val
}

func (p *Postgres) GetMetric(metricName string, metricType string) (*repository.Metric, bool) {
	stmt := p.GetSelectStmtByMetricType(metricType)
	row := stmt.QueryRowContext(context.Background(), metricName)
	output, err := p.ScanMetricByMetricType(row, metricType)
	return output, err == nil
}

func (p *Postgres) GetAllMetricsByType(metricType string) []repository.Metric {
	output := make([]repository.Metric, 0)
	tableName := ""
	switch metricType {
	case repository.CounterMetricKey:
		tableName = CounterMetricsTableName
	case repository.GaugeMetricKey:
		tableName = GaugeMetricsTableName
	}
	rows, err := p.db.QueryContext(context.Background(), "SELECT * FROM "+tableName)
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

func (p *Postgres) PingRepo() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := p.db.PingContext(ctx); err != nil {
		return false
	}

	return true
}

func (p *Postgres) CloseRepository() {
	p.db.Close()
}

func (p *Postgres) GetSelectStmtByMetricType(metricType string) *sql.Stmt {
	switch metricType {
	case repository.CounterMetricKey:
		return p.selectCounterMetricStmt
	case repository.GaugeMetricKey:
		return p.selectGaugeMetricStmt
	default:
		return nil
	}
}

func (p *Postgres) ScanMetricByMetricType(row *sql.Row, metricType string) (output *repository.Metric, err error) {
	output = &repository.Metric{}
	switch metricType {
	case repository.CounterMetricKey:
		err = row.Scan(&output.Delta)
		if err != nil {
			return
		}
	case repository.GaugeMetricKey:
		err = row.Scan(&output.Value)
		if err != nil {
			return
		}
	}
	err = row.Err()
	return
}
