#! /usr/bin/env bash
#! /usr/bin/env bash
go version
echo $USER
echo $PATH
source /etc/profile
go version
echo $USER
echo $PATH
# Step3: go e2e test
ginkgo -timeout=10h --fail-fast  --label-filter=${E2E_TESTING_LEVEL} test/e2e

