#!/bin/bash

results=$( make providerlint 2>&1 )
while [[ "${results}" == *Killed* ]]; do
  echo "${results}"
  results=$( make providerlint 2>&1 )
done
echo "${results}"

echo "==> Checking source code with providerlint..."
providerlint \
  -c 1 \
  -AT001.ignored-filename-suffixes=_data_source_test.go \
  -AWSAT006=false \
  -AWSR002=false \
  -AWSV001=false \
  -R001=false \
  -R010=false \
  -R018=false \
  -R019=false \
  -V001=false \
  -V009=false \
  -V011=false \
  -V012=false \
  -V013=false \
  -V014=false \
  -XR001=false \
  -XR002=false \
  -XR003=false \
  -XR004=false \
  -XR005=false \
  -XS001=false \
  -XS002=false \
  ./internal/service/... \
  ./internal/provider/...