package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type Postgres struct {
	db *sql.DB
}

const (
	GaugeMetricsTableName   = "gauge_metrics"
	CounterMetricsTableName = "counter_metrics"
)

func NewPostgresRepo(adress string) (*Postgres, error) {
	db, err := sql.Open("pgx", adress)
	if err != nil {
		return nil, err
	}

	err = MigrateTable(db, GaugeMetricsTableName, "double precision")
	if err != nil {
		return nil, err
	}

	err = MigrateTable(db, CounterMetricsTableName, "BIGINT")
	if err != nil {
		return nil, err
	}

	instance := Postgres{
		db: db,
	}
	return &instance, nil
}

func MigrateTable(db *sql.DB, tableName string, valueType string) error {
	_, err := db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS "+tableName+
		"(metric_id varchar(150) NOT NULL, metric_value "+valueType+" NOT NULL)")
	if err != nil {
		logger.Log.Error(err)
		return err
	}

	return nil
}

func (p *Postgres) UpdateMetric(ctx context.Context, metric *repository.Metric) *repository.Metric {
	switch metric.MType {
	case repository.GaugeMetricKey:
		return p.UpdateGaugeMetric(ctx, metric)
	case repository.CounterMetricKey:
		return p.UpdateCounterMetricValue(ctx, metric)
	}
	return nil
}

func (p *Postgres) UpdateMetrics(ctx context.Context, metrics []repository.Metric) ([]repository.Metric, error) {
	output := metrics
	var err error
	retry(3, 1*time.Second, func() error {
		output, err = p.UpdateWithRetries(ctx, metrics)
		if err != nil && isRetriableError(err) {
			return err
		}
		return nil
	})
	return output, err
}

func retry(attempts int, delay time.Duration, operation func() error) error {
	for i := 0; i < attempts; i++ {
		err := operation()
		if err != nil {
			time.Sleep(delay)
			delay += 2 * time.Second
			continue
		}
		return err
	}
	return fmt.Errorf("retry failed after %d attempts", attempts)
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)
}

func (p *Postgres) UpdateWithRetries(ctx context.Context, metrics []repository.Metric) ([]repository.Metric, error) {
	output := make([]repository.Metric, 0)
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, metric := range metrics {
		var ouputMetric *repository.Metric
		switch metric.MType {
		case repository.GaugeMetricKey:
			ouputMetric, err = p.UpdateGaugeMetricWithTx(ctx, tx, &metric)
		case repository.CounterMetricKey:
			ouputMetric, err = p.UpdateCounterMetricValueWithTx(ctx, tx, &metric)
		}

		if err != nil {
			return nil, err
		} else {
			output = append(output, *ouputMetric)
		}
	}

	return output, nil
}

func (p *Postgres) UpdateGaugeMetric(ctx context.Context, metric *repository.Metric) *repository.Metric {
	res, err := p.db.ExecContext(ctx, "UPDATE "+GaugeMetricsTableName+` 
	SET metric_value = $1 WHERE metric_id = $2`, metric.Value, metric.ID)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		_, err = p.db.ExecContext(ctx, "INSERT INTO "+GaugeMetricsTableName+` 
		(metric_id, metric_value) 
		VALUES ($1, $2)`, metric.ID, metric.Value)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	}

	return metric
}

func (p *Postgres) UpdateGaugeMetricWithTx(ctx context.Context, tx *sql.Tx, metric *repository.Metric) (*repository.Metric, error) {
	res, err := tx.ExecContext(ctx, "UPDATE "+GaugeMetricsTableName+` 
	SET metric_value = $1 WHERE metric_id = $2`, metric.Value, metric.ID)
	if err != nil {
		logger.Log.Error(err)
		return nil, err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		_, err = tx.ExecContext(ctx, "INSERT INTO "+GaugeMetricsTableName+` 
		(metric_id, metric_value) 
		VALUES ($1, $2)`, metric.ID, metric.Value)
		if err != nil {
			logger.Log.Error(err)
			return nil, err
		}
	}

	return metric, nil
}

func (p *Postgres) UpdateCounterMetricValue(ctx context.Context, metric *repository.Metric) *repository.Metric {
	val, ok := p.GetMetric(ctx, metric.ID, metric.MType)

	if !ok {
		val = metric
		_, err := p.db.ExecContext(ctx, "INSERT INTO "+CounterMetricsTableName+" (metric_id, metric_value) VALUES ($1, $2)", val.ID, *val.Delta)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	} else {
		newDelta := (*metric.Delta) + (*val.Delta)
		val.Delta = &newDelta
		_, err := p.db.ExecContext(ctx, "UPDATE "+CounterMetricsTableName+" SET metric_value = $1 WHERE metric_id = $2", *val.Delta, val.ID)
		if err != nil {
			logger.Log.Error(err)
			return nil
		}
	}

	return val
}

func (p *Postgres) UpdateCounterMetricValueWithTx(ctx context.Context, tx *sql.Tx, metric *repository.Metric) (*repository.Metric, error) {
	val, ok := p.GetMetricWithTX(ctx, tx, metric.ID, metric.MType, CounterMetricsTableName)

	if !ok {
		val = metric
		_, err := tx.ExecContext(ctx, "INSERT INTO "+CounterMetricsTableName+" (metric_id, metric_value) VALUES ($1, $2)", val.ID, val.Delta)
		if err != nil {
			logger.Log.Error(err)
			return nil, err
		}
	} else {
		newDelta := (*metric.Delta) + (*val.Delta)
		val.Delta = &newDelta
		_, err := tx.ExecContext(ctx, "UPDATE "+CounterMetricsTableName+" SET metric_value = $1 WHERE metric_id = $2", val.Delta, val.ID)
		if err != nil {
			logger.Log.Error(err)
			return nil, err
		}
	}

	return val, nil
}

func (p *Postgres) GetMetricWithTX(ctx context.Context, tx *sql.Tx, metricName string, metricType string, metricTableName string) (*repository.Metric, bool) {
	row := tx.QueryRowContext(ctx, "SELECT metric_value FROM "+metricTableName+" WHERE metric_id = $1", metricName)
	output, err := p.ScanMetricByMetricType(row, metricType)
	output.ID = metricName
	output.MType = metricType
	return output, err == nil
}

func (p *Postgres) GetMetric(ctx context.Context, metricName string, metricType string) (*repository.Metric, bool) {
	var reuqestStmt *sql.Stmt
	switch metricType {
	case repository.GaugeMetricKey:
		stmt, err := p.db.PrepareContext(ctx, "SELECT metric_value FROM "+GaugeMetricsTableName+" WHERE metric_id = $1")
		if err != nil {
			logger.Log.Error(err)
			return nil, false
		}
		reuqestStmt = stmt
	case repository.CounterMetricKey:
		stmt, err := p.db.PrepareContext(ctx, "SELECT metric_value FROM "+CounterMetricsTableName+" WHERE metric_id = $1")
		if err != nil {
			logger.Log.Error(err)
			return nil, false
		}
		reuqestStmt = stmt
	}
	row := reuqestStmt.QueryRowContext(ctx, metricName)
	output, err := p.ScanMetricByMetricType(row, metricType)
	output.ID = metricName
	output.MType = metricType
	reuqestStmt.Close()
	return output, err == nil
}

func (p *Postgres) GetAllMetricsByType(ctx context.Context, metricType string) []repository.Metric {
	output := make([]repository.Metric, 0)
	tableName := ""
	switch metricType {
	case repository.CounterMetricKey:
		tableName = CounterMetricsTableName
	case repository.GaugeMetricKey:
		tableName = GaugeMetricsTableName
	}
	rows, err := p.db.QueryContext(ctx, "SELECT * FROM "+tableName)
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
		case repository.CounterMetricKey:
			rows.Scan(&metric.ID, &metric.Delta)
			if err != nil {
				logger.Log.Error(err)
				return output
			}
		}

		output = append(output, metric)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error(err)
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
