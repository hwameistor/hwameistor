#! /usr/bin/env bash

git clone https://github.com/hwameistor/helm-charts.git
cd  ./e2e && ginkgo --fail-fast