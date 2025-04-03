package types

import (
	"errors"
)

var ErrUnsupportedMetricType error = errors.New("unsupported metric type")
var ErrCantFindMetric error = errors.New("can't find metric")
var ErrMetricNilValue error = errors.New("value for update metric is nill")
var ErrUnsupportedMetricValueType error = errors.New("unsupported value type")
