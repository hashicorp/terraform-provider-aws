#!/bin/bash

echo "==> Checking source code with providerlint..."

function do_provider_linting() {
  args=("$@")

  local paths=""

  for (( i=0;i<${#args[@]};i++ )); do
    if [ $i -ne 0 ]; then
      paths+=" "
    fi
    paths+="./internal/service/${args[${i}]}/..."
  done

  echo "Packages: ${paths}"

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
    ./internal/provider/... \
    "${paths}"
}

do_provider_linting $@
