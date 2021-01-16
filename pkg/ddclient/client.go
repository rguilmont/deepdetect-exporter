package ddclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout = 15 * time.Second
)

type ServiceStatisticsFromDD struct {
	Body struct {
		ServiceStats struct {
			PredictSuccess         *int     `json:"predict_success"`
			InferenceCount         *int     `json:"inference_count"`
			PredictFailure         *int     `json:"predict_failure"`
			TotalPredictDuration   *float64 `json:"total_predict_duration_ms"`
			PredictCount           *int     `json:"predict_count"`
			TotalTransformDuration *float64 `json:"total_transform_duration_ms"`
			AvgBatchSize           *float64 `json:"avg_batch_size"`
		} `json:"service_stats"`
		Type       string
		MLType     string
		Repository string
		Stats      struct {
			DataMemTest  *int `json:"data_mem_test"`
			DataMemTrain *int `json:"data_mem_train"`
			Flops        *int `json:"flops"`
			Params       *int `json:"params"`
		} `json:"stats"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Mllib       string `json:"mllib"`
		Predict     bool   `json:"predict"`
	} `json:"body"`
}

// DDInfo is a minimalistic info endpoint catching just the needed fields for exporter
type DDInfo struct {
	Head struct {
		Commit   string `json:"commit"`
		Version  string `json:"version"`
		Services []struct {
			Name string `json:"name"`
		} `json:"services"`
	} `json:"head"`
}

type DDMetricsClient struct {
	httpClient *http.Client
	endpoint   url.URL
}

func NewDDMetricsClient(endpoint url.URL) DDMetricsClient {
	client := DDMetricsClient{
		http.DefaultClient,
		endpoint,
	}

	client.httpClient.Timeout = defaultTimeout
	return client
}

// GetInfo returns information about DeepDetect instance
func (client DDMetricsClient) GetInfo() (*DDInfo, error) {
	infoURL := client.endpoint
	infoURL.Path = "/info"
	resp, err := client.httpClient.Get(infoURL.String())

	if err != nil {
		return nil, err
	}

	info := DDInfo{}
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}
	logrus.Debugln("Received /info from Deepdetect :", info)
	return &info, nil
}

// GetMetrics returns a list of ServiceStatisticsFromDD, or an error
func (client DDMetricsClient) GetMetrics() ([]ServiceStatisticsFromDD, error) {

	info, err := client.GetInfo()
	if err != nil {
		return nil, err
	}
	metrics := []ServiceStatisticsFromDD{}
	for _, s := range info.Head.Services {
		statsURL := client.endpoint
		statsURL.Path = fmt.Sprintf("/services/%v", s.Name)

		resp, err := client.httpClient.Get(statsURL.String())
		stats := ServiceStatisticsFromDD{}
		err = json.NewDecoder(resp.Body).Decode(&stats)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, stats)
	}

	return metrics, nil
}
