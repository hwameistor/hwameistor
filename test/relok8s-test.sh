#! /usr/bin/env bash

set -x
set -e

relok8s chart move helm/hwameistor/ --image-patterns helm/hwameistor/.relok8s-images.yaml  --registry $ImageRegistry2 --repo-prefix hwameistorex -y
