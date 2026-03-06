// SPDX-License-Identifier: MIT

package environment

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/czerwonk/junos_exporter/pkg/collector"
)

const prefix string = "junos_environment_"

var (
	statusValues = map[string]int{
		"OK":      1,
		"Testing": 2,
		"Failed":  3,
		"Absent":  4,
		"Present": 5,
	}

	temperaturesDesc *prometheus.Desc
	powerSupplyDesc  *prometheus.Desc
	fanStatusDesc    *prometheus.Desc
	fanAirflowDesc   *prometheus.Desc
	pemDesc          *prometheus.Desc
	fanDesc          *prometheus.Desc
	dcVoltageDesc    *prometheus.Desc
	dcCurrentDesc    *prometheus.Desc
	dcPowerDesc      *prometheus.Desc
	dcLoadDesc       *prometheus.Desc
	dcOutputDesc     *prometheus.Desc
	inputVoltageDesc *prometheus.Desc
	inputCurrentDesc *prometheus.Desc
	inputPowerDesc   *prometheus.Desc
)

func init() {
	l := []string{"target", "re_name", "item"}
	temperaturesDesc = prometheus.NewDesc(prefix+"item_temp", "Temperature of the air flowing past", l, nil)
	powerSupplyDesc = prometheus.NewDesc(prefix+"power_up", "Status of power supplies (1 OK, 2 Testing, 3 Failed, 4 Absent, 5 Present)", append(l, "status"), nil)
	fanStatusDesc = prometheus.NewDesc(prefix+"fan_up", "Status of fans (1 OK, 2 Testing, 3 Failed, 4 Absent, 5 Present)", append(l, "status"), nil)
	fanAirflowDesc = prometheus.NewDesc(prefix+"fan_airflow_up", "Status of	fan airflows (1 OK, 2 Testing, 3 Failed, 4 Absent, 5 Present)", append(l, "status"), nil)

	pemDesc = prometheus.NewDesc(prefix+"pem_state", "State of PEM module. 1 - Online, 2 - Present, 3 - Empty", append(l, "state"), nil)
	dcVoltageDesc = prometheus.NewDesc(prefix+"pem_voltage", "PEM voltage value", l, nil)
	dcCurrentDesc = prometheus.NewDesc(prefix+"pem_current", "PEM current value", l, nil)
	dcPowerDesc = prometheus.NewDesc(prefix+"pem_power_usage", "PEM power usage in W", l, nil)
	dcLoadDesc = prometheus.NewDesc(prefix+"pem_power_load_percent", "PEM power usage percent of total", l, nil)
	dcOutputDesc = prometheus.NewDesc(prefix+"pem_dc_output", "PSM DC output status (1 OK, 0 not OK)", l, nil)
	inputVoltageDesc = prometheus.NewDesc(prefix+"pem_input_voltage", "PSU input voltage in V", l, nil)
	inputCurrentDesc = prometheus.NewDesc(prefix+"pem_input_current", "PSU input current in A", l, nil)
	inputPowerDesc = prometheus.NewDesc(prefix+"pem_input_power", "PSU input power in W", l, nil)

	l = []string{"target", "re_name", "item", "fan_name"}
	fanDesc = prometheus.NewDesc(prefix+"pem_fanspeed", "Fan speed in RPM", l, nil)
}

type environmentCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &environmentCollector{}
}

// Name returns the name of the collector
func (*environmentCollector) Name() string {
	return "Environment"
}

// Describe describes the metrics
func (*environmentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- temperaturesDesc
	ch <- fanDesc
	ch <- dcPowerDesc
	ch <- dcOutputDesc
	ch <- inputVoltageDesc
	ch <- inputCurrentDesc
	ch <- inputPowerDesc
}

// deviceModel returns the lowercase product model string for the device.
// Most devices expose it via 'show version', but some (e.g. EX4300) return
// chassis-inventory instead, so we fall back to 'show chassis hardware'.
func (c *environmentCollector) deviceModel(client collector.Client) (string, error) {
	var v showVersionResult
	if err := client.RunCommandAndParse("show version", &v); err != nil {
		return "", errors.Wrap(err, "failed to run command 'show version'")
	}

	if model := strings.ToLower(v.SoftwareInformation.ProductModel); model != "" {
		return model, nil
	}

	// Fallback: EX4300 returns chassis-inventory instead of software-information
	var hw showChassisHardwareResult
	if err := client.RunCommandAndParse("show chassis hardware", &hw); err == nil {
		return strings.ToLower(hw.ChassisInventory.Chassis.Description), nil
	}

	return "", nil
}

// Collect collects metrics from JunOS
func (c *environmentCollector) Collect(client collector.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	model, err := c.deviceModel(client)
	if err != nil {
		return err
	}

	c.environmentItems(client, ch, labelValues)

	if strings.Contains(model, "qfx5220") {
		c.environmentPEMItemsQFX5220(client, ch, labelValues)
	} else if strings.Contains(model, "ex4300") {
		c.environmentPEMItemsEX4300(client, ch, labelValues)
	} else {
		c.environmentPEMItems(client, ch, labelValues)
	}
	return nil
}

func (c *environmentCollector) environmentItems(client collector.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	x := multiEngineResult{}


	err := client.RunCommandAndParseWithParser("show chassis environment", func(b []byte) error {
		return parseXML(b, &x)
	})
	if err != nil {
		return nil
	}

	if client.IsSatelliteEnabled() {
		var y = multiEngineResult{}
		err = client.RunCommandAndParseWithParser("show chassis environment satellite", func(b []byte) error {
			if string(b[:]) == "\nerror: syntax error, expecting <command>: satellite\n" {
				log.Printf("system doesn't seem to have satellite enabled")
				return nil
			}

			return parseXML(b, &y)
		})
		if err != nil {
			return nil
		}

		if len(y.Results.RoutingEngines) > 0 {
			x.Results.RoutingEngines[0].EnvironmentInformation.Items = append(x.Results.RoutingEngines[0].EnvironmentInformation.Items, y.Results.RoutingEngines[0].EnvironmentInformation.Items...)
		}
	}

	for _, re := range x.Results.RoutingEngines {
		l := labelValues
		for _, item := range re.EnvironmentInformation.Items {
			l = append(labelValues, re.Name)
			if containsAny(item.Name, []string{"Power Supply", "PEM", "PSM"}) {
				l = append(l, item.Name, item.Status)
				ch <- prometheus.MustNewConstMetric(powerSupplyDesc, prometheus.GaugeValue, float64(statusValues[item.Status]), l...)
			} else if strings.Contains(item.Name, "Fan") {
				l = append(l, item.Name, item.Status)
				if strings.Contains(item.Name, "Airflow") {
					ch <- prometheus.MustNewConstMetric(fanAirflowDesc, prometheus.GaugeValue, float64(statusValues[item.Status]), l...)
				} else {
					ch <- prometheus.MustNewConstMetric(fanStatusDesc, prometheus.GaugeValue, float64(statusValues[item.Status]), l...)
				}
			} else if item.Temperature != nil {
				l = append(l, item.Name)
				ch <- prometheus.MustNewConstMetric(temperaturesDesc, prometheus.GaugeValue, item.Temperature.Value, l...)
			}
		}
	}

	return nil
}

func (c *environmentCollector) environmentPEMItems(client collector.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var x = multiEngineResult{}

	stateValues := map[string]int{
		"Online":  1,
		"Present": 2,
		"Empty":   3,
		"Offline": 4,
	}

	err := client.RunCommandAndParseWithParser("show chassis environment pem", func(b []byte) error {
		return parseXML(b, &x)
	})
	if err != nil {
		err := client.RunCommandAndParseWithParser("show chassis environment psm", func(b []byte) error {
			return parseXML(b, &x)
		})
		if err != nil {
			return err
		}
	}

	for _, re := range x.Results.RoutingEngines {
		for _, e := range re.EnvironmentComponentInformation.EnvironmentComponentItem {
			l := append(labelValues, re.Name, e.Name)

			ch <- prometheus.MustNewConstMetric(pemDesc, prometheus.GaugeValue, float64(stateValues[e.State]), append(l, e.State)...)

			for _, f := range e.FanSpeedReading {
				rpms, err := strconv.ParseFloat(strings.TrimSuffix(f.FanSpeed, " RPM"), 64)
				if err != nil {
					return fmt.Errorf("could not parse fan speed value to float: %s", f.FanSpeed)
				}
				ch <- prometheus.MustNewConstMetric(fanDesc, prometheus.GaugeValue, rpms, append(l, f.FanName)...)
			}

			voltage := 0.0
			if e.DcInformation.DcDetail.DcVoltage > 0 {
				voltage = e.DcInformation.DcDetail.DcVoltage
			}

			if e.DcInformation.DcDetail.Str3DcVoltage > 0 {

				voltage = e.DcInformation.DcDetail.Str3DcVoltage
			}

			if voltage > 0 {
				ch <- prometheus.MustNewConstMetric(dcVoltageDesc, prometheus.GaugeValue, voltage, l...)
				ch <- prometheus.MustNewConstMetric(dcCurrentDesc, prometheus.GaugeValue, e.DcInformation.DcDetail.DcCurrent, l...)
				ch <- prometheus.MustNewConstMetric(dcPowerDesc, prometheus.GaugeValue, e.DcInformation.DcDetail.DcPower, l...)
				ch <- prometheus.MustNewConstMetric(dcLoadDesc, prometheus.GaugeValue, e.DcInformation.DcDetail.DcLoad, l...)
			}
		}
	}

	return nil
}


func (c *environmentCollector) environmentPEMItemsQFX5220(client collector.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	x := environmentPEMResultModelQFX5220{}

	stateValues := map[string]int{
		"Online":  1,
		"Present": 2,
		"Empty":   3,
		"Offline": 4,
	}

	err := client.RunCommandAndParseWithParser("show chassis environment pem", func(b []byte) error {
		return xml.Unmarshal(b, &x)
	})
	if err != nil {
		err := client.RunCommandAndParseWithParser("show chassis environment psm", func(b []byte) error {
			return xml.Unmarshal(b, &x)
		})
		if err != nil {
			return err
		}
	}

	reName := "N/A"
	for _, item := range x.EnvironmentComponentInformation.EnvironmentComponentItem {
		l := append(labelValues, reName, item.Name)

		ch <- prometheus.MustNewConstMetric(pemDesc, prometheus.GaugeValue, float64(stateValues[item.State]), append(l, item.State)...)

		fan1Speed := item.PsmInformation.FanSpeedReadingPsm.Fan1Speed
		if fan1Speed != "" {
			rpms, err := strconv.ParseFloat(strings.TrimSuffix(fan1Speed, " RPM"), 64)
			if err != nil {
				return fmt.Errorf("could not parse fan speed value to float: %s", fan1Speed)
			}
			ch <- prometheus.MustNewConstMetric(fanDesc, prometheus.GaugeValue, rpms, append(l, item.PsmInformation.FanSpeedReadingPsm.Fan1Name)...)
		}

		//it could be that the DCOutputValue has the same states as stateValues from above
		//but I couldn't verify it for sure
		dcOutputVal := 0.0
		if strings.EqualFold(strings.ToLower(item.PsmInformation.PsmStatus.DcOutput), "ok") {
			dcOutputVal = 1.0
		}
		ch <- prometheus.MustNewConstMetric(dcOutputDesc, prometheus.GaugeValue, dcOutputVal, l...)
	}

	return nil
}


func (c *environmentCollector) environmentPEMItemsEX4300(client collector.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	x := environmentPEMResultModelEX4300{}

	stateValues := map[string]int{
		"Online":  1,
		"Present": 2,
		"Empty":   3,
		"Offline": 4,
	}

	err := client.RunCommandAndParseWithParser("show chassis environment power-supply-unit", func(b []byte) error {
		return xml.Unmarshal(b, &x)
	})
	if err != nil {
		return err
	}

	reName := "N/A"
	for _, item := range x.EnvironmentComponentInformation.EnvironmentComponentItem {
		pem := item.PemInformation
		itemName := fmt.Sprintf("FPC %s PSU %s", pem.FpcSlot, pem.PemSlot)
		l := append(labelValues, reName, itemName)

		ch <- prometheus.MustNewConstMetric(pemDesc, prometheus.GaugeValue, float64(stateValues[pem.PemState]), append(l, pem.PemState)...)
		ch <- prometheus.MustNewConstMetric(temperaturesDesc, prometheus.GaugeValue, pem.PemTemperature, append(labelValues, reName, itemName)...)
		ch <- prometheus.MustNewConstMetric(dcVoltageDesc, prometheus.GaugeValue, pem.OutputVolt, l...)
		ch <- prometheus.MustNewConstMetric(dcCurrentDesc, prometheus.GaugeValue, pem.OutputCurrent, l...)
		ch <- prometheus.MustNewConstMetric(dcPowerDesc, prometheus.GaugeValue, pem.OutputPower, l...)
		ch <- prometheus.MustNewConstMetric(inputVoltageDesc, prometheus.GaugeValue, pem.InputVolt, l...)
		ch <- prometheus.MustNewConstMetric(inputCurrentDesc, prometheus.GaugeValue, pem.InputCurrent, l...)
		ch <- prometheus.MustNewConstMetric(inputPowerDesc, prometheus.GaugeValue, pem.InputPower, l...)
	}

	return nil
}

func containsAny(s string, items []string) bool {
	for _, item := range items {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}

func parseXML(b []byte, res *multiEngineResult) error {
	if strings.Contains(string(b), "multi-routing-engine-results") {
		return xml.Unmarshal(b, res)
	}

	fi := singleEngineResult{}

	err := xml.Unmarshal(b, &fi)
	if err != nil {
		return err
	}

	res.Results.RoutingEngines = []routingEngine{
		{
			Name:                            "N/A",
			EnvironmentComponentInformation: fi.EnvironmentComponentInformation,
			EnvironmentInformation:          fi.EnvironmentInformation,
		},
	}
	return nil
}
