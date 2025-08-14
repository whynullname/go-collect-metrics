package grpcserver

import (
	"context"
	"errors"
	"net"

	"github.com/whynullname/go-collect-metrics/internal/logger"
	pb "github.com/whynullname/go-collect-metrics/internal/proto"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/types"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	pb.UnimplementedMetricsServer

	metricsUseCase *metrics.MetricsUseCase
}

func NewGrpcServer(useCase *metrics.MetricsUseCase) *GrpcServer {
	return &GrpcServer{
		metricsUseCase: useCase,
	}
}

func (g *GrpcServer) ListenServer() error {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterMetricsServer(s, g)

	if err := s.Serve(listen); err != nil {
		return err
	}

	return nil
}

func (g *GrpcServer) GetAllMetrics(ctx context.Context, in *pb.GetAllMetricsRequest) (*pb.GetAllMetricsResponse, error) {
	var response pb.GetAllMetricsResponse
	gaugeMetrics, err := g.metricsUseCase.GetAllMetricsByType(ctx, repository.GaugeMetricKey)
	if err != nil {
		response.Error = types.ErrInternalError.Error()
		return nil, err
	}
	counterMetrics, err := g.metricsUseCase.GetAllMetricsByType(ctx, repository.CounterMetricKey)
	if err != nil {
		response.Error = types.ErrInternalError.Error()
		return nil, err
	}

	response.GaugeMetrics = make([]*pb.Metric, len(gaugeMetrics))
	for _, metric := range gaugeMetrics {
		newOutputMetric := &pb.Metric{
			Id:    metric.ID,
			Type:  metric.MType,
			Delta: *metric.Delta,
			Value: *metric.Value,
		}

		response.GaugeMetrics = append(response.GaugeMetrics, newOutputMetric)
	}

	response.CounterMetrics = make([]*pb.Metric, len(gaugeMetrics))
	for _, metric := range counterMetrics {
		newOutputMetric := &pb.Metric{
			Id:    metric.ID,
			Type:  metric.MType,
			Delta: *metric.Delta,
			Value: *metric.Value,
		}

		response.CounterMetrics = append(response.CounterMetrics, newOutputMetric)
	}

	return &response, nil
}

func (g *GrpcServer) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	var response pb.UpdateMetricResponse
	requestMetric := &repository.Metric{
		MType: in.Metric.Type,
		ID:    in.Metric.Id,
	}

	updatedMetric, err := g.metricsUseCase.UpdateMetric(ctx, requestMetric)
	if err != nil {
		logger.Log.Errorf("Error with update metrics: %w", err)
		if errors.Is(err, types.ErrMetricNilValue) || errors.Is(err, types.ErrUnsupportedMetricType) {
			response.Error = err.Error()
			return nil, err
		}

		response.Error = types.ErrInternalError.Error()
		return nil, err
	}

	response.UpdatedMetric.Id = updatedMetric.ID
	response.UpdatedMetric.Type = updatedMetric.MType
	response.UpdatedMetric.Delta = *updatedMetric.Delta
	response.UpdatedMetric.Value = *updatedMetric.Value

	return &response, nil
}

func (g *GrpcServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var response pb.UpdateMetricsResponse
	metrics := make([]repository.Metric, len(in.RequestMetrics))

	for _, inputMetric := range in.RequestMetrics {
		metric := repository.Metric{
			ID:    inputMetric.Id,
			MType: inputMetric.Type,
		}

		metrics = append(metrics, metric)
	}

	updatedMetrics, err := g.metricsUseCase.UpdateMetrics(ctx, metrics)
	if err != nil {
		logger.Log.Errorf("Error with update metrics: %w", err)

		if errors.Is(err, types.ErrMetricNilValue) || errors.Is(err, types.ErrUnsupportedMetricType) {
			response.Error = err.Error()
			return nil, err
		}

		response.Error = types.ErrInternalError.Error()
		return nil, err
	}

	response.UpdatedMetrics = make([]*pb.Metric, len(updatedMetrics))
	for _, metric := range updatedMetrics {
		outputMetric := &pb.Metric{
			Id:    metric.ID,
			Type:  metric.MType,
			Delta: *metric.Delta,
			Value: *metric.Value,
		}

		response.UpdatedMetrics = append(response.UpdatedMetrics, outputMetric)
	}

	return &response, nil
}

func (g *GrpcServer) GetMetricByName(ctx context.Context, in *pb.GetMetricByNameRequest) (*pb.GetMetricByNameResponse, error) {
	var response pb.GetMetricByNameResponse

	val, err := g.metricsUseCase.GetMetric(ctx, in.Type, in.Name)
	if err != nil {
		response.Error = err.Error()
		return nil, err
	}

	response.Metric = &pb.Metric{
		Id:    val.ID,
		Type:  val.MType,
		Delta: *val.Delta,
		Value: *val.Value,
	}

	return &response, nil
}
