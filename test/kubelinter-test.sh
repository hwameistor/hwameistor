#! /usr/bin/env bash

set -x


if ! command -v kube-linter &> /dev/null
then
  echo "kube-linter could not be found"
  wget https://github.com/stackrox/kube-linter/releases/download/v0.7.2/kube-linter-linux.tar.gz
  tar -zxvf kube-linter-linux.tar.gz
  sudo chmod +x kube-linter-linux
  sudo cp kube-linter /usr/local/bin/kube-linter
else
  echo "kube-linter is installed"
fi

if ! command -v jq &> /dev/null
then
  echo "jq could not be found"
  wget https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux-amd64
  sudo chmod +x jq-linux-amd64
  sudo cp jq-linux-amd64 /usr/local/bin/jq
else
  echo "jq is installed"
fi


time=$(date +%Y%m%d%H%M)
body=$(kube-linter lint ./helm/hwameistor/ --do-not-auto-add-defaults --include "unset-cpu-requirements,unset-memory-requirements")


JSON_DATA=$(jq -n \
  --arg title "Kube-Linter auto Issue $time" \
  --arg body "\`\`\`$body" \
  --argjson labels '["kind/bug"]' \
  '{title: $title, body: $body, labels: $labels}')

curl -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${token}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/hwameistor/hwameistor/issues \
  -d "$JSON_DATA"