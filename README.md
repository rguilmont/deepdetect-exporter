# DeepDetect Prometheus exporter

Prometheus exporter for various metrics about DeepDetect, written in Go.

# Installation

## Docker

```
docker pull rguilmont/deepdetect-exporter:v0.12
docker run --name deepdetect-exporter -p 8181:8181 rguilmont/deepdetect-exporter:v0.12 -listen 0.0.0.0:8181 -monitor http://deepdetect:8080
```