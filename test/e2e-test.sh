#! /usr/bin/env bash

git clone https://github.com/hwameistor/helm-charts.git test/helm-charts
ginkgo --fail-fast test/e2e