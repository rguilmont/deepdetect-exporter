package ddclient

import (
	"net/url"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type DeepDetectCollector struct {
	PredictSuccess       *prometheus.Desc
	PredictFailure       *prometheus.Desc
	InferenceCount       *prometheus.Desc
	PredictDurationSum   *prometheus.Desc
	TransformDurationSum *prometheus.Desc
	PredictCount         *prometheus.Desc
	AvgBatchSize         *prometheus.Desc
	DataMemTest          *prometheus.Desc
	DataMemTrain         *prometheus.Desc
	Flops                *prometheus.Desc
	Params               *prometheus.Desc

	// Other metrics for DD availability
	Available *prometheus.Desc
	cli       *DDMetricsClient
}

func NewDeepDetectCollector(endpoint url.URL) (*DeepDetectCollector, error) {

	cli := NewDDMetricsClient(endpoint)

	info, err := cli.GetInfo()
	if err != nil {
		return nil, err
	}
	constLabels := prometheus.Labels{
		"dd_version": info.Head.Version,
		"dd_commit":  info.Head.Commit,
	}
	return &DeepDetectCollector{
		PredictSuccess:       prometheus.NewDesc("dd_predict_success_count", "PredictSuccess value", labelsInOrder(), constLabels),
		PredictFailure:       prometheus.NewDesc("dd_predict_failure_count", "PredictFailure value", labelsInOrder(), constLabels),
		InferenceCount:       prometheus.NewDesc("dd_inference_count", "InferenceCount value", labelsInOrder(), constLabels),
		PredictCount:         prometheus.NewDesc("dd_predict_count", "PredictCount value", labelsInOrder(), constLabels),
		PredictDurationSum:   prometheus.NewDesc("dd_predict_duration_sum", "Total prediction time in ms", labelsInOrder(), constLabels),
		TransformDurationSum: prometheus.NewDesc("dd_transform_duration_sum", "Total ", labelsInOrder(), constLabels),
		AvgBatchSize:         prometheus.NewDesc("dd_batch_size_avg", "AvgBatchSize value", labelsInOrder(), constLabels),
		DataMemTest:          prometheus.NewDesc("dd_data_mem_test", "DataMemTest value", labelsInOrder(), constLabels),
		DataMemTrain:         prometheus.NewDesc("dd_data_mem_train", "DataMemTrain value", labelsInOrder(), constLabels),
		Flops:                prometheus.NewDesc("dd_flops", "Flops value", labelsInOrder(), constLabels),
		Params:               prometheus.NewDesc("dd_params", "Params value", labelsInOrder(), constLabels),

		Available: prometheus.NewDesc("dd_available", "DeepDetect availability", nil, constLabels),
		cli:       &cli,
	}, nil

}

func (c *DeepDetectCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.PredictSuccess
	ch <- c.PredictFailure
	ch <- c.PredictCount
	ch <- c.PredictDurationSum
	ch <- c.TransformDurationSum
	ch <- c.PredictCount
	ch <- c.AvgBatchSize
	ch <- c.DataMemTest
	ch <- c.DataMemTrain
	ch <- c.Flops
	ch <- c.Params
}

// maybeMetric Send metric, if exist, to chanel. If value of metric is nil, or uncastable to float64, then print a warning or an error.
func (st *ServiceStatisticsFromDD) maybeMetric(ch chan<- prometheus.Metric, p *prometheus.Desc, valueType prometheus.ValueType, value interface{}) {

	if reflect.ValueOf(value).IsNil() {
		// It happens because DD api response is still evolving.
		//  i'll add versioning of response later.
		logrus.Warnf("Expected metric %v cannot be found in DeepDetect response. Skipping.", p.String())
		return
	}

	logrus.Info(value)
	var prometheusValue float64

	switch vt := value.(type) {
	case *int:
		prometheusValue = float64(*vt)
	case *float64:
		prometheusValue = *vt
	default:
		logrus.Errorf("Convertion metric %v from type %v to float ( prometheus standard ) unknown. Skipping.", p.String(), reflect.TypeOf(vt))
		return
	}

	ch <- prometheus.MustNewConstMetric(
		p,
		valueType,
		prometheusValue,
		evaluateLabels(*st)...,
	)
}

func (c *DeepDetectCollector) Collect(ch chan<- prometheus.Metric) {
	metrics, err := c.cli.GetMetrics()

	if err != nil {
		logrus.Errorln("Error while scraping DeepDetect metrics:", err)
		ch <- prometheus.NewInvalidMetric(c.Available, err)
	}

	for _, serviceMetrics := range metrics {

		serviceMetrics.maybeMetric(ch, c.AvgBatchSize, prometheus.GaugeValue, serviceMetrics.Body.ServiceStats.AvgBatchSize)
		serviceMetrics.maybeMetric(ch, c.PredictSuccess, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.PredictSuccess)
		serviceMetrics.maybeMetric(ch, c.PredictFailure, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.PredictFailure)
		serviceMetrics.maybeMetric(ch, c.PredictCount, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.PredictCount)
		serviceMetrics.maybeMetric(ch, c.InferenceCount, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.InferenceCount)
		serviceMetrics.maybeMetric(ch, c.PredictDurationSum, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.TotalPredictDuration)
		serviceMetrics.maybeMetric(ch, c.TransformDurationSum, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.TotalTransformDuration)
		serviceMetrics.maybeMetric(ch, c.DataMemTest, prometheus.GaugeValue, serviceMetrics.Body.Stats.DataMemTest)
		serviceMetrics.maybeMetric(ch, c.DataMemTrain, prometheus.GaugeValue, serviceMetrics.Body.Stats.DataMemTrain)
		serviceMetrics.maybeMetric(ch, c.Flops, prometheus.GaugeValue, serviceMetrics.Body.Stats.Flops)
		serviceMetrics.maybeMetric(ch, c.Params, prometheus.GaugeValue, serviceMetrics.Body.Stats.Params)

	}

	ch <- prometheus.MustNewConstMetric(c.Available, prometheus.GaugeValue, 1.0)
}
