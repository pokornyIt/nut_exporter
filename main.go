package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"time"
)

const (
	nameSpace       = "nut"
	applicationName = "nut_exporter"
)

var (
	logger    log.Logger // logger
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
)

func updateStatus(output string) {
	upsStatusValue := upsStatusRegex.FindAllStringSubmatch(output, -1)[0][1]

	switch upsStatusValue {
	case "CAL":
		upsStatus.Set(0)
	case "TRIM":
		upsStatus.Set(1)
	case "BOOST":
		upsStatus.Set(2)
	case "OL":
		upsStatus.Set(3)
	case "OB":
		upsStatus.Set(4)
	case "OVER":
		upsStatus.Set(5)
	case "LB":
		upsStatus.Set(6)
	case "RB":
		upsStatus.Set(7)
	case "BYPASS":
		upsStatus.Set(8)
	case "OFF":
		upsStatus.Set(9)
	case "CHRG":
		upsStatus.Set(10)
	case "DISCHRG":
		upsStatus.Set(11)
	}
}

func readVarList(conn connection) string {
	err := conn.open()
	if err != nil {
		return ""
	}
	defer conn.close()
	data, err := conn.getList("VAR")
	if err != nil {
		_ = level.Error(logger).Log("msg", err)
		return ""
	}
	return data
}

func recordMetrics() {
	for _, metric := range metricsList {
		prometheus.MustRegister(metric.metrics)
	}
	for _, metric := range metricsVecList {
		prometheus.MustRegister(metric.metrics)
	}
	prometheus.MustRegister(upsStatus)
	_ = level.Debug(logger).Log("msg", "create connection for NUT server", "host", config.getServer())
	connection := *newConnection(config.getServer(), config.User, config.Password, config.UpsName)

	go func() {
		for {
			upsOutput := readVarList(connection)

			if len(upsOutput) == 0 {
				_ = level.Error(logger).Log("msg", "problem read data from NUT server")
			} else {
				for _, metric := range metricsList {
					metric.updateFromSource(upsOutput)
				}
				for _, metric := range metricsVecList {
					metric.updateFromSource(upsOutput)
				}
				updateStatus(upsOutput)
			}
			time.Sleep(time.Duration(config.Refresh) * time.Second)
		}
	}()
}

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	version.Branch = Branch
	version.Revision = Revision
	version.BuildUser = BuildUser
	version.BuildDate = BuildDate
	version.Version = Version
	kingpin.Version(version.Print(applicationName))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger = promlog.New(promlogConfig)
	_ = level.Info(logger).Log("msg", "Starting NUT exporter on ups "+config.UpsName, "version", version.Info())

	err := config.loadFile(*configFile)

	if *showConfig {
		_ = level.Info(logger).Log("msg", "show only configuration ane exit")
		fmt.Print(config.print())
		os.Exit(0)
	}
	if err != nil {
		_ = level.Error(logger).Log("msg", "problem with configuration", "error", err)
		fmt.Printf("Program did not start due to configuration error! \r\n\tError: %s", err)
		os.Exit(1)
	}

	_ = level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	http.Handle("/metrics", promhttp.Handler())

	recordMetrics()

	_ = level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	_ = http.ListenAndServe(*listenAddress, nil)
}
