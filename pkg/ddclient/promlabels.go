package ddclient

type labelsEvaluator []struct {
	label     string
	evaluator func(ServiceStatisticsFromDD) string
}

var orderedLabels labelsEvaluator = []struct {
	label     string
	evaluator func(ServiceStatisticsFromDD) string
}{
	{
		"service_name",
		func(s ServiceStatisticsFromDD) string {
			return s.Body.Name
		},
	},
	{
		"repository",
		func(s ServiceStatisticsFromDD) string {
			return s.Body.Repository
		},
	},
	{
		"type",
		func(s ServiceStatisticsFromDD) string {
			return s.Body.Type
		},
	},
	{
		"ml_type",
		func(s ServiceStatisticsFromDD) string {
			return s.Body.MLType
		},
	},
	{
		"service_description",
		func(s ServiceStatisticsFromDD) string {
			return s.Body.Description
		},
	},
	{
		"predict",
		func(s ServiceStatisticsFromDD) string {
			if s.Body.Predict {
				return "1"
			}
			return "0"
		},
	},
}

func labelsInOrder() []string {
	res := []string{}
	for _, e := range orderedLabels {
		res = append(res, e.label)
	}
	return res
}

func evaluateLabels(s ServiceStatisticsFromDD) []string {
	res := []string{}
	for _, e := range orderedLabels {
		res = append(res, e.evaluator(s))
	}
	return res
}
