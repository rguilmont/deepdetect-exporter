package ddclient

// metricTransformer allows to transform the format of a metric ( such as converting ms to seconds )
type metricTransformer func(float64) float64

func noopTransformer(v float64) float64 {
	return v
}

func msToSec(v float64) float64 {
	return v / 1000
}
