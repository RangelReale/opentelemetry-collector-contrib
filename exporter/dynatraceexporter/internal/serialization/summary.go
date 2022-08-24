package serialization

import (
	dtMetric "github.com/dynatrace-oss/dynatrace-metric-utils-go/metric"
	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/dimensions"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func serializeSummaryPoint(name, prefix string, dims dimensions.NormalizedDimensionList, dp pmetric.SummaryDataPoint) (string, error) {
	if dp.Count() == 0 {
		return "", nil
	}

	min, max, sum := summaryDataPointToSummary(dp)

	dm, err := dtMetric.NewMetric(
		name,
		dtMetric.WithPrefix(prefix),
		dtMetric.WithDimensions(dims),
		dtMetric.WithTimestamp(dp.Timestamp().AsTime()),
		dtMetric.WithFloatSummaryValue(min, max, sum, int64(dp.Count())),
	)

	if err != nil {
		return "", err
	}

	return dm.Serialize()
}

func serializeSummary(logger *zap.Logger, prefix string, metric pmetric.Metric, defaultDimensions dimensions.NormalizedDimensionList, staticDimensions dimensions.NormalizedDimensionList, metricLines []string) []string {
	summary := metric.Summary()

	for i := 0; i < summary.DataPoints().Len(); i++ {
		dp := summary.DataPoints().At(i)
		line, err := serializeSummaryPoint(
			metric.Name(),
			prefix,
			makeCombinedDimensions(defaultDimensions, dp.Attributes(), staticDimensions),
			dp,
		)

		if err != nil {
			logger.Warn(
				"Error serializing summary data point",
				zap.String("name", metric.Name()),
				zap.Error(err),
			)
		}

		if line != "" {
			metricLines = append(metricLines, line)
		}
	}
	return metricLines
}

func summaryDataPointToSummary(dp pmetric.SummaryDataPoint) (float64, float64, float64) {
	if dp.QuantileValues().Len() == 0 {
		return 0, 0, dp.Sum()
	}

	min, max := dp.QuantileValues().At(0).Value(), dp.QuantileValues().At(0).Value()
	for bi := 1; bi < dp.QuantileValues().Len(); bi++ {
		value := dp.QuantileValues().At(bi).Value()
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max, dp.Sum()
}
