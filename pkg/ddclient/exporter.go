package ddclient

import (
	"net/url"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type DeepDetectCollector struct {
	PredictRequestsSuccessTotal *prometheus.Desc
	PredictRequestsFailureTotal *prometheus.Desc
	InferenceRequestsTotal      *prometheus.Desc
	PredictDurationTotal        *prometheus.Desc
	TransformDurationTotal      *prometheus.Desc
	PredictRequestsTotal        *prometheus.Desc
	AvgBatchSize                *prometheus.Desc
	DataMemTest                 *prometheus.Desc
	DataMemTrain                *prometheus.Desc
	Flops                       *prometheus.Desc
	Params                      *prometheus.Desc

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
		PredictRequestsSuccessTotal: prometheus.NewDesc("deepdetect_predict_requests_success_total", "Total number of succesful predicts", labelsInOrder(), constLabels),
		PredictRequestsFailureTotal: prometheus.NewDesc("deepdetect_predict_requests_failure_total", "Total number of failed predicts", labelsInOrder(), constLabels),
		InferenceRequestsTotal:      prometheus.NewDesc("deepdetect_inference_requests_total", "Total number of succesful inferences", labelsInOrder(), constLabels),
		PredictRequestsTotal:        prometheus.NewDesc("deepdetect_predict_requests_total", "Total number of predicts", labelsInOrder(), constLabels),
		PredictDurationTotal:        prometheus.NewDesc("deepdetect_predict_duration_seconds_total", "Total prediction time in seconds", labelsInOrder(), constLabels),
		TransformDurationTotal:      prometheus.NewDesc("deepdetect_transform_duration_seconds_total", "Total transformation time in seconds", labelsInOrder(), constLabels),
		AvgBatchSize:                prometheus.NewDesc("deepdetect_batch_size_avg", "AvgBatchSize value", labelsInOrder(), constLabels),
		DataMemTest:                 prometheus.NewDesc("deepdetect_data_mem_test_bytes", "DataMemTest value", labelsInOrder(), constLabels),
		DataMemTrain:                prometheus.NewDesc("deepdetect_data_mem_train_bytes", "DataMemTrain value", labelsInOrder(), constLabels),
		Flops:                       prometheus.NewDesc("deepdetect_flops", "Flops value", labelsInOrder(), constLabels),
		Params:                      prometheus.NewDesc("deepdetect_params", "Params value", labelsInOrder(), constLabels),

		Available: prometheus.NewDesc("dd_available", "DeepDetect availability", nil, constLabels),
		cli:       &cli,
	}, nil

}

func (c *DeepDetectCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.PredictRequestsSuccessTotal
	ch <- c.PredictRequestsFailureTotal
	ch <- c.PredictRequestsTotal
	ch <- c.PredictDurationTotal
	ch <- c.TransformDurationTotal
	ch <- c.AvgBatchSize
	ch <- c.DataMemTest
	ch <- c.DataMemTrain
	ch <- c.Flops
	ch <- c.Params
}

// maybeMetric Send metric, if exist, to chanel. If value of metric is nil, or uncastable to float64, then print a warning or an error.
// Transform func will transform
func (st *ServiceStatisticsFromDD) maybeMetric(ch chan<- prometheus.Metric, p *prometheus.Desc, valueType prometheus.ValueType, value interface{}, transformFunc metricTransformer) {

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
		transformFunc(prometheusValue),
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

		serviceMetrics.maybeMetric(ch, c.AvgBatchSize, prometheus.GaugeValue, serviceMetrics.Body.ServiceStats.AvgBatchSize, noopTransformer)
		serviceMetrics.maybeMetric(ch, c.PredictRequestsSuccessTotal, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.PredictSuccess, noopTransformer)
		serviceMetrics.maybeMetric(ch, c.PredictRequestsFailureTotal, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.PredictFailure, noopTransformer)
		serviceMetrics.maybeMetric(ch, c.PredictRequestsTotal, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.PredictCount, noopTransformer)
		serviceMetrics.maybeMetric(ch, c.InferenceRequestsTotal, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.InferenceCount, noopTransformer)
		serviceMetrics.maybeMetric(ch, c.PredictDurationTotal, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.TotalPredictDuration, msToSec)
		serviceMetrics.maybeMetric(ch, c.TransformDurationTotal, prometheus.CounterValue, serviceMetrics.Body.ServiceStats.TotalTransformDuration, msToSec)
		serviceMetrics.maybeMetric(ch, c.DataMemTest, prometheus.GaugeValue, serviceMetrics.Body.Stats.DataMemTest, noopTransformer)
		serviceMetrics.maybeMetric(ch, c.DataMemTrain, prometheus.GaugeValue, serviceMetrics.Body.Stats.DataMemTrain, noopTransformer)
		serviceMetrics.maybeMetric(ch, c.Flops, prometheus.GaugeValue, serviceMetrics.Body.Stats.Flops, noopTransformer)
		serviceMetrics.maybeMetric(ch, c.Params, prometheus.GaugeValue, serviceMetrics.Body.Stats.Params, noopTransformer)

	}

	ch <- prometheus.MustNewConstMetric(c.Available, prometheus.GaugeValue, 1.0)
}
