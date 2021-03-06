package vcentermultiplier

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
)

type VCenterMultiplier struct {
	Config      []string
	VerboseMode bool

	isInitialized        bool
	array                map[string]map[string]float64
	totalMhz_average     float64
	effectivecpu_average float64
}

var sampleConfig = `
  ## Config can contain multiply factors for each metrics.
  ## Each config line should be the string in influx format.
  Config = [
    "mem used_percent=100,available_percent=100",
    "swap used_percent=100"
  ]

  # VerboseMode allows to print changes for debug purpose
  VerboseMode = false

  #totalmhz_average and effectivecpu_average will be used to 
  #compute effectivecpu_average as % instead of Mhz
}
`

func (VCenterMultiplier *VCenterMultiplier) SampleConfig() string {
	return sampleConfig
}

func (VCenterMultiplier *VCenterMultiplier) Description() string {
	return "Multiply metrics values on some multiply factor."
}

func (VCenterMultiplier *VCenterMultiplier) Apply(metricsArray ...telegraf.Metric) []telegraf.Metric {
	if !VCenterMultiplier.isInitialized {
		VCenterMultiplier.Initialize()
		VCenterMultiplier.isInitialized = true
	}

	// Loop for all metrics
	for i, metrics := range metricsArray {

		if val, ok := metrics.Fields()["totalmhz_average"]; ok == true {
			VCenterMultiplier.totalMhz_average = toFloat(val)
		}
		if val, ok := metrics.Fields()["effectivecpu_average"]; ok == true {
			VCenterMultiplier.effectivecpu_average = toFloat(val)
		}

		if _, ok := VCenterMultiplier.array[metrics.Name()]; ok == true {

			newFields := make(map[string]interface{})

			// Loop for specified metric
			for metricName, metricValue := range metrics.Fields() {

				newValue := metricValue

				// Check that current metric should be multiplied
				if factor, ok := VCenterMultiplier.array[metrics.Name()][metricName]; ok == true {
					newValue = VCenterMultiplier.Multiply(metricValue, factor)
					log.Printf("[processor.VCenterMultiplier] [%v.%v] %v * %v => %v\n",
						metrics.Name(),
						metricName,
						metricValue,
						factor,
						newValue)
				}

				newFields[metricName] = newValue
			}

			newMetric, err := metric.New(metrics.Name(),
				metrics.Tags(), newFields, metrics.Time(), metrics.Type())

			if err != nil {
				log.Printf("[processor.VCenterMultiplier] Cannot make a copy: %v\n", err)
			} else {
				metricsArray[i] = newMetric
			}
		}
	}
	// Update the metricsArray to hold effectivecpu_average as % instead of Mhz
	if VCenterMultiplier.totalMhz_average != 0 {
		cpu_available := int((VCenterMultiplier.effectivecpu_average * 100) / VCenterMultiplier.totalMhz_average)
		for i, metrics := range metricsArray {
			if _, ok := metrics.Fields()["effectivecpu_average"]; ok == true {
				newFields := make(map[string]interface{})

				for metricName, metricValue := range metrics.Fields() {
					newValue := metricValue

					if metricName == "effectivecpu_average" {
						newValue = cpu_available
					}

					newFields[metricName] = newValue
				}
				newMetric, err := metric.New(metrics.Name(), metrics.Tags(), newFields, metrics.Time(), metrics.Type())
				if err == nil {
					metricsArray[i] = newMetric
				}
			}
		}
	}

	return metricsArray
}

func (VCenterMultiplier *VCenterMultiplier) Multiply(value interface{}, factor float64) interface{} {
	switch data := value.(type) {
	case int:
		return int(factor * float64(data))
	case uint:
		return uint(factor * float64(data))
	case int32:
		return int32(factor * float64(data))
	case uint32:
		return uint32(factor * float64(data))
	case int64:
		return int64(factor * float64(data))
	case uint64:
		return uint64(factor * float64(data))
	case float32:
		return float32(factor * float64(data))
	case float64:
		return float64(factor * float64(data))
	default:
		log.Printf("[processor.VCenterMultiplier] can not multiply %v [float64] with value: %T '%v'\n", factor, value, data)
	}
	return value
}

func toFloat(value interface{}) float64 {
	switch data := value.(type) {
	case int:
		return float64(data)
	case int32:
		return float64(data)
	case int64:
		return float64(data)
	case float32:
		return float64(data)
	case float64:
		return data
	default:
		log.Printf("[processor.VCenterMultiplier] plugin couldn't create 'float64' from value: %T '%v'\n", value, data)
	}
	return 0
}

func (VCenterMultiplier *VCenterMultiplier) Initialize() error {

	VCenterMultiplier.totalMhz_average = 0
	VCenterMultiplier.effectivecpu_average = 0
	VCenterMultiplier.array = make(map[string]map[string]float64)

	for _, str := range VCenterMultiplier.Config {
		parser, _ := parsers.NewInfluxParser()
		metrics, err := parser.ParseLine(str)
		if err != nil {
			log.Printf("E! %v\n", err)
			continue
		}

		keeper, ok := VCenterMultiplier.array[metrics.Name()]
		if !ok {
			keeper = make(map[string]float64)
			VCenterMultiplier.array[metrics.Name()] = keeper
		}

		for metricName, _metricValue := range metrics.Fields() {
			metricValue := toFloat(_metricValue)
			keeper[metricName] = metricValue
		}
	}

	return nil
}

func init() {
	processors.Add("vcentermultiplier", func() telegraf.Processor {
		return &VCenterMultiplier{}
	})
}
