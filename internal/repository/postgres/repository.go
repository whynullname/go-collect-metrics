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
	"github.com/whynullname/go-collect-metrics/internal/repository/types"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgresRepo(adress string) (*Postgres, error) {
	db, err := sql.Open("pgx", adress)
	if err != nil {
		return nil, err
	}

	err = MigrateTable(db, "gauge_metrics", "double precision")
	if err != nil {
		return nil, err
	}

	err = MigrateTable(db, "counter_metrics", "BIGINT")
	if err != nil {
		return nil, err
	}

	instance := Postgres{
		db: db,
	}
	return &instance, nil
}

func MigrateTable(db *sql.DB, tableName string, valueType string) error {
	_, err := db.ExecContext(context.TODO(), "CREATE TABLE IF NOT EXISTS "+tableName+
		"(metric_id varchar(150) NOT NULL, metric_value "+valueType+" NOT NULL)")
	if err != nil {
		logger.Log.Error(err)
		return err
	}

	return nil
}

func (p *Postgres) UpdateMetric(ctx context.Context, metric *repository.Metric) (*repository.Metric, error) {
	switch metric.MType {
	case repository.GaugeMetricKey:
		return p.UpdateGaugeMetric(ctx, metric)
	case repository.CounterMetricKey:
		return p.UpdateCounterMetricValue(ctx, metric)
	}
	return nil, types.ErrUnsupportedMetricType
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

func (p *Postgres) UpdateGaugeMetric(ctx context.Context, metric *repository.Metric) (*repository.Metric, error) {
	res, err := p.db.ExecContext(ctx, `UPDATE gauge_metrics
	SET metric_value = $1 WHERE metric_id = $2`, metric.Value, metric.ID)
	if err != nil {
		return nil, err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		_, err = p.db.ExecContext(ctx, `INSERT INTO gauge_metrics 
		(metric_id, metric_value) 
		VALUES ($1, $2)`, metric.ID, metric.Value)
		if err != nil {
			return nil, err
		}
	}

	return metric, nil
}

func (p *Postgres) UpdateGaugeMetricWithTx(ctx context.Context, tx *sql.Tx, metric *repository.Metric) (*repository.Metric, error) {
	res, err := tx.ExecContext(ctx, `UPDATE gauge_metrics 
	SET metric_value = $1 WHERE metric_id = $2`, metric.Value, metric.ID)
	if err != nil {
		return nil, err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO gauge_metrics 
		(metric_id, metric_value) 
		VALUES ($1, $2)`, metric.ID, metric.Value)
		if err != nil {
			return nil, err
		}
	}

	return metric, nil
}

func (p *Postgres) UpdateCounterMetricValue(ctx context.Context, metric *repository.Metric) (*repository.Metric, error) {
	val, err := p.GetMetric(ctx, metric.ID, metric.MType)

	if err != nil {
		val = metric
		_, err := p.db.ExecContext(ctx, "INSERT INTO counter_metrics (metric_id, metric_value) VALUES ($1, $2)", val.ID, *val.Delta)
		if err != nil {
			return nil, err
		}
	} else {
		newDelta := metric.GetDelta() + val.GetDelta()
		val.Delta = &newDelta
		_, err := p.db.ExecContext(ctx, "UPDATE counter_metrics SET metric_value = $1 WHERE metric_id = $2", *val.Delta, val.ID)
		if err != nil {
			return nil, err
		}
	}

	return val, nil
}

func (p *Postgres) UpdateCounterMetricValueWithTx(ctx context.Context, tx *sql.Tx, metric *repository.Metric) (*repository.Metric, error) {
	val, err := p.GetMetricWithTX(ctx, tx, metric.ID, metric.MType, "counter_metrics")

	if err != nil {
		val = metric
		_, err := tx.ExecContext(ctx, "INSERT INTO counter_metrics (metric_id, metric_value) VALUES ($1, $2)", val.ID, val.Delta)
		if err != nil {
			return nil, err
		}
	} else {
		newDelta := metric.GetDelta() + val.GetDelta()
		val.Delta = &newDelta
		_, err := tx.ExecContext(ctx, "UPDATE counter_metrics SET metric_value = $1 WHERE metric_id = $2", val.Delta, val.ID)
		if err != nil {
			return nil, err
		}
	}

	return val, nil
}

func (p *Postgres) GetMetricWithTX(ctx context.Context, tx *sql.Tx, metricName string, metricType string, metricTableName string) (*repository.Metric, error) {
	row := tx.QueryRowContext(ctx, "SELECT metric_value FROM "+metricTableName+" WHERE metric_id = $1", metricName)
	output, err := p.ScanMetricByMetricType(row, metricType)
	output.ID = metricName
	output.MType = metricType
	return output, err
}

func (p *Postgres) GetMetric(ctx context.Context, metricName string, metricType string) (*repository.Metric, error) {
	switch metricType {
	case repository.GaugeMetricKey:
		return p.GetMetricQurey(ctx, metricName, metricType, "gauge_metrics")
	case repository.CounterMetricKey:
		return p.GetMetricQurey(ctx, metricName, metricType, "counter_metrics")
	}

	return nil, types.ErrUnsupportedMetricType
}

func (p *Postgres) GetMetricQurey(ctx context.Context, metricName string, metricType string, metricTableName string) (*repository.Metric, error) {
	row := p.db.QueryRowContext(ctx, "SELECT metric_value FROM "+metricTableName+" WHERE metric_id = $1", metricName)
	output, err := p.ScanMetricByMetricType(row, metricType)
	if err != nil {
		return nil, types.ErrCantFindMetric
	}
	output.ID = metricName
	output.MType = metricType
	return output, err
}

func (p *Postgres) GetAllMetricsByType(ctx context.Context, metricType string) ([]repository.Metric, error) {
	output := make([]repository.Metric, 0)
	tableName := ""
	switch metricType {
	case repository.CounterMetricKey:
		tableName = "counter_metrics"
	case repository.GaugeMetricKey:
		tableName = "gauge_metrics"
	}
	rows, err := p.db.QueryContext(ctx, "SELECT * FROM "+tableName)
	if err != nil {
		return output, err
	}

	for rows.Next() {
		metric := repository.Metric{MType: metricType}
		switch metricType {
		case repository.GaugeMetricKey:
			err = rows.Scan(&metric.ID, &metric.Value)
			if err != nil {
				return output, err
			}
		case repository.CounterMetricKey:
			rows.Scan(&metric.ID, &metric.Delta)
			if err != nil {
				return output, err
			}
		}

		output = append(output, metric)
	}

	err = rows.Err()
	return output, err
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
