package vcentermultiplier

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

type testData struct {
	testName string
	vcm      *VCenterMultiplier
	metrics  []telegraf.Metric
	expected []telegraf.Metric
}

func TestVCenterNoChange(t *testing.T) {
	now := time.Now()
	test := testData{
		testName: "noChange",
		vcm:      &VCenterMultiplier{},
		metrics: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"name": "idle_time",
				},
				map[string]interface{}{
					"value": int(128),
				},
				now,
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"name": "idle_time",
				},
				map[string]interface{}{
					"value": int(128),
				},
				now,
			),
		},
	}
	t.Run(test.testName, func(t *testing.T) {
		actual := test.vcm.Apply(test.metrics...)
		testutil.RequireMetricsEqual(t, test.expected, actual)
	})
}

func TestVCenterMultiplyBy10(t *testing.T) {
	now := time.Now()
	test := testData{
		testName: "multiply10",
		vcm: &VCenterMultiplier{
			isInitialized: true,
			VerboseMode:   true,
			array: map[string]map[string]float64{
				"cpu": map[string]float64{"value": 10},
			},
			totalMhz_average: 0,
		},
		metrics: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"name": "idle_time",
				},
				map[string]interface{}{
					"value": int(128),
				},
				now,
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"name": "idle_time",
				},
				map[string]interface{}{
					"value": int(1280),
				},
				now,
			),
		},
	}
	t.Run(test.testName, func(t *testing.T) {
		actual := test.vcm.Apply(test.metrics...)
		testutil.RequireMetricsEqual(t, test.expected, actual)
	})
}

func TestVCenterEffectivecpuAverage(t *testing.T) {
	now := time.Now()
	test := testData{
		testName: "EffectivecpuAverage(",
		vcm: &VCenterMultiplier{
			isInitialized: true,
			VerboseMode:   true,
			array: map[string]map[string]float64{
				"cpu": map[string]float64{"value": 10},
			},
			totalMhz_average:     0,
			effectivecpu_average: 0,
		},
		metrics: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"name": "tag1",
				},
				map[string]interface{}{
					"effectivecpu_average": int(100),
				},
				now,
			),
			testutil.MustMetric("cpu",
				map[string]string{
					"name": "tag2",
				},
				map[string]interface{}{
					"totalmhz_average": int(200),
				},
				now,
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"name": "tag1",
				},
				map[string]interface{}{
					"effectivecpu_average": int(50),
				},
				now,
			),
			testutil.MustMetric("cpu",
				map[string]string{
					"name": "tag2",
				},
				map[string]interface{}{
					"totalmhz_average": int(200),
				},
				now,
			)},
	}

	t.Run(test.testName, func(t *testing.T) {
		actual := test.vcm.Apply(test.metrics...)
		testutil.RequireMetricsEqual(t, test.expected, actual)
	})
}
