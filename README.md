# DeepDetect Prometheus exporter

Prometheus exporter for various metrics about DeepDetect, written in Go. The goal of this exporter is to hit a deepdetect process directly.
If you run deepdetect in kubernetes, it would fit perfectly as a sidecar pod.

# Installation

## Docker

```
docker pull rguilmont/deepdetect-exporter:v0.1
docker run --name deepdetect-exporter -p 8181:8181 rguilmont/deepdetect-exporter:v0.1 -listen 0.0.0.0:8181 -monitor http://deepdetect:8080
```

# Available metrics

| deepdetect_predict_requests_success_total   | Total number of succesful predicts               |
|---------------------------------------------|--------------------------------------------------|
| deepdetect_predict_requests_failure_total   | Total number of failed predicts                  |
| deepdetect_inference_requests_total         | Total number of succesful inferences             |
| deepdetect_predict_requests_total           | Total number of predicts                         |
| deepdetect_predict_duration_seconds_total   | Total prediction time in seconds                 |
| deepdetect_transform_duration_seconds_total | Total transformation time in seconds             |
| deepdetect_batch_size_avg                   | AvgBatchSize value                               |
| deepdetect_data_mem_test_bytes              | DataMemTest value                                |
| deepdetect_data_mem_train_bytes             | DataMemTrain value                               |
| deepdetect_flops                            | Flops value"                                     |
| deepdetect_params                           | Params value                                     |
| dd_available                                | DeepDetect process status (1 for up, 0 for down) |

## Labels

Each metrics but `dd_available` depends on loaded services. So they have the following labels exposed: `service_name`, `repository`, `type`, `ml_type`, `service_description`, `predict`.

Also exposed for each metrics, labels giving informations on running deepdetect process: `dd_version` and `dd_commit`.

# Grafana dashboard

There's a grafana dashboard JSON file available in the `grafana` dashboard that you can import.
