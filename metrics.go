package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"strconv"
)

type metricsGauge struct {
	metrics  prometheus.Gauge
	analyzer *regexp.Regexp
}

type metricsGaugeVec struct {
	metrics  *prometheus.GaugeVec
	analyzer *regexp.Regexp
	name     string
}

type metricFunc interface {
	updateFromSource(output string)
}

// Regex
var (
	batteryChargeRegex         = regexp.MustCompile(`(?:battery[.]charge:(?:\s)(.*))`)
	batteryChargeLowRegex      = regexp.MustCompile(`(?:battery[.]charge[.]low:(?:\s)(.*))`)
	batteryChargeWarnRegex     = regexp.MustCompile(`(?:battery[.]charge[.]warning:(?:\s)(.*))`)
	batteryPacksRegex          = regexp.MustCompile(`(?:battery[.]packs:(?:\s)(.*))`)
	batteryTypeRegex           = regexp.MustCompile(`(?:battery[.]type:(?:\s)(.*))`)
	batteryVoltageRegex        = regexp.MustCompile(`(?:battery[.]voltage:(?:\s)(.*))`)
	batteryVoltageNominalRegex = regexp.MustCompile(`(?:battery[.]voltage[.]nominal:(?:\s)(.*))`)
	deviceMFRRegex             = regexp.MustCompile(`(?:device[.]mfr:(?:\s)(.*))`)
	deviceModelRegex           = regexp.MustCompile(`(?:device[.]model:(?:\s)(.*))`)
	deviceTypeRegex            = regexp.MustCompile(`(?:device[.]type:(?:\s)(.*))`)
	driverNameRegex            = regexp.MustCompile(`(?:driver[.]name:(?:\s)(.*))`)
	//driverPoolFreqRegex        = regexp.MustCompile(`(?:driver[.]parameter[.]poolfreq:(?:\s)(.*))`)
	//driverPoolIntervalRegex    = regexp.MustCompile(`(?:driver[.]parameter[.]poolinterval:(?:\s)(.*))`)
	driverVersionRegex        = regexp.MustCompile(`(?:driver[.]version:(?:\s)(.*))`)
	driverVersionDataRegex    = regexp.MustCompile(`(?:driver[.]version[.]data:(?:\s)(.*))`)
	inputVoltageRegex         = regexp.MustCompile(`(?:input[.]voltage:(?:\s)(.*))`)
	inputVoltageNominalRegex  = regexp.MustCompile(`(?:input[.]voltage[.]nominal:(?:\s)(.*))`)
	outputVoltageRegex        = regexp.MustCompile(`(?:output[.]voltage:(?:\s)(.*))`)
	outputVoltageNominalRegex = regexp.MustCompile(`(?:output[.]voltage[.]nominal:(?:\s)(.*))`)
	upsBeeperStatusRegex      = regexp.MustCompile(`(?:ups[.]beeper[.]status:(?:\s)(.*))`)
	upsDelayShutRegex         = regexp.MustCompile(`(?:ups[.]delay[.]shutdown:(?:\s)(.*))`)
	upsDelayStartRegex        = regexp.MustCompile(`(?:ups[.]delay[.]start:(?:\s)(.*))`)
	upsLoadRegex              = regexp.MustCompile(`(?:ups[.]load:(?:\s)(.*))`)
	upsMFRRegex               = regexp.MustCompile(`(?:ups[.]mfr:(?:\s)(.*))`)
	upsModelRegex             = regexp.MustCompile(`(?:ups[.]model:(?:\s)(.*))`)
	upsPowerNominalRegex      = regexp.MustCompile(`(?:ups[.]power[.]nominal:(?:\s)(.*))`)
	upsRealPowerNominalRegex  = regexp.MustCompile(`(?:ups[.]realpower[.]nominal:(?:\s)(.*))`)
	upsTempRegex              = regexp.MustCompile(`(?:ups[.]temperature:(?:\s)(.*))`)
	upsStatusRegex            = regexp.MustCompile(`(?:ups[.]status:(?:\s)(.*))`)
)

// NUT Gauges https://networkupstools.org/docs/user-manual.chunked/apcs01.html
var (
	batteryCharge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "battery_charge",
		Help:      "Current battery charge (percent)",
	})

	batteryChargeLow = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "battery_charge_low",
		Help:      "Remaining battery level when UPS switches to LB state (percent)",
	})

	batteryChargeWarning = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "battery_charge_warning",
		Help:      "Battery level when UPS switches to \"Warning\" state (percent)",
	})

	batteryPacks = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "battery_pack",
		Help:      "Number of battery packs on the UPS",
	})

	batteryType = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "battery_type",
		Help:      "Battery chemistry",
	}, []string{"type"})

	batteryVoltage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "battery_voltage",
		Help:      "Current battery voltage",
	})

	batteryVoltageNominal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "battery_voltage_nominal",
		Help:      "Nominal battery voltage",
	})

	deviceMfr = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "device_mfr",
		Help:      "Device manufacturer",
	}, []string{"manufacturer"})

	deviceModel = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "device_model",
		Help:      "Device model",
	}, []string{"model"})

	deviceType = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "device_type",
		Help:      "Device type (ups, pdu, scd, psu, ats)",
	}, []string{"type"})

	driverName = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "driver_name",
		Help:      "Driver name",
	}, []string{"name"})

	driverVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "driver_version",
		Help:      "Driver version (NUT release)",
	}, []string{"version"})

	driverVersionData = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "driver_version_data",
		Help:      "Version of the internal data mapping, for generic drivers",
	}, []string{"data"})

	inputVoltage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "input_voltage",
		Help:      "Current input voltage",
	})

	inputVoltageNominal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "input_voltage_nominal",
		Help:      "Nominal input voltage",
	})

	outputVoltage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "output_voltage",
		Help:      "Current output voltage",
	})

	outputVoltageNominal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "output_voltage_nominal",
		Help:      "Nominal output voltage",
	})

	upsBeeperStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_beeper_status",
		Help:      "UPS beeper status (enabled, disabled or muted)",
	}, []string{"status"})

	upsDelayShut = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_delay_shutdown",
		Help:      "Interval to wait after shutdown with delay command (seconds)",
	})

	upsDelayStart = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_delay_start",
		Help:      "Interval to wait before restarting the load (seconds)",
	})

	upsLoad = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_load",
		Help:      "Current UPS load (percent)",
	})

	upsMfr = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_mfr",
		Help:      "UPS manufacturer",
	}, []string{"manufacturer"})

	upsModel = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_model",
		Help:      "UPS model",
	}, []string{"model"})

	upsPowerNominal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_power_nominal",
		Help:      "Nominal value of apparent power (Volt-Amps)",
	})

	upsRealPowerNominal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_real_power_nominal",
		Help:      "Nominal value of real power (Watts)",
	})

	upsTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_temp",
		Help:      "UPS Temperature (degrees C)",
	})

	upsStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: NameSpace,
		Name:      "ups_status",
		Help:      "Current UPS Status (0=Calibration, 1=SmartTrim, 2=SmartBoost, 3=Online, 4=OnBattery, 5=Overloaded, 6=LowBattery, 7=ReplaceBattery, 8=OnBypass, 9=Off, 10=Charging, 11=Discharging)",
	})
)

var metricsList = []metricsGauge{
	{batteryCharge, batteryChargeRegex},
	{batteryChargeLow, batteryChargeLowRegex},
	{batteryChargeWarning, batteryChargeWarnRegex},
	{batteryPacks, batteryPacksRegex},
	{batteryVoltage, batteryVoltageRegex},
	{batteryVoltageNominal, batteryVoltageNominalRegex},
	{inputVoltage, inputVoltageRegex},
	{inputVoltageNominal, inputVoltageNominalRegex},
	{outputVoltage, outputVoltageRegex},
	{outputVoltageNominal, outputVoltageNominalRegex},
	{upsDelayShut, upsDelayShutRegex},
	{upsDelayStart, upsDelayStartRegex},
	{upsLoad, upsLoadRegex},
	{upsPowerNominal, upsPowerNominalRegex},
	{upsRealPowerNominal, upsRealPowerNominalRegex},
	{upsTemp, upsTempRegex},
}
var metricsVecList = []metricsGaugeVec{
	{batteryType, batteryTypeRegex, "type"},
	{deviceMfr, deviceMFRRegex, "manufacturer"},
	{deviceModel, deviceModelRegex, "model"},
	{deviceType, deviceTypeRegex, "type"},
	{driverName, driverNameRegex, "name"},
	{driverVersion, driverVersionRegex, "version"},
	{driverVersionData, driverVersionDataRegex, "data"},
	{upsBeeperStatus, upsBeeperStatusRegex, "status"},
	{upsMfr, upsMFRRegex, "manufacturer"},
	{upsModel, upsModelRegex, "model"},
}

func (gauge *metricsGauge) updateFromSource(output string) {
	if gauge.analyzer.FindAllStringSubmatch(output, -1) == nil {
		prometheus.Unregister(gauge.metrics)
	} else {
		getData, _ := strconv.ParseFloat(gauge.analyzer.FindAllStringSubmatch(output, -1)[0][1], 64)
		gauge.metrics.Set(getData)
	}
}

func (gaugeVec *metricsGaugeVec) updateFromSource(output string) {
	if gaugeVec.analyzer.FindAllStringSubmatch(output, -1) == nil {
		prometheus.Unregister(gaugeVec.metrics)
	} else {
		getData := gaugeVec.analyzer.FindAllStringSubmatch(output, -1)[0][1]
		gaugeVec.metrics.With(prometheus.Labels{gaugeVec.name: getData}).Set(1)
	}
}
