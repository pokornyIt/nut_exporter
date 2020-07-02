![Build Status](https://github.com/pokornyIt/nut_exporter/workflows/Build/badge.svg)
[![GitHub](https://img.shields.io/github/license/pokornyIt/nut_exporter)](/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/pokornyIt/nut_exporter)](https://goreportcard.com/report/github.com/pokornyIt/nut_exporter)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/pokornyit/nut_exporter?label=latest)](https://github.com/pokornyIt/nut_exporter/releases/latest)


# NUT Exporter
Network UPS Tools (NUT) exporter for Prometheus based on [p404/nut_exporter](https://github.com/p404/nut_exporter) and extend it.
NUT client inspired by [lzap/gonutclient](https://github.com/lzap/gonutclient).

# What extend
- Export all possible data
- Improve logging
- Add more config options
- Direct communicate with NUT server

# Not support
- secure connection (for now)
