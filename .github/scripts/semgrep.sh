#!/bin/bash

local idx=$1

local results=$( semgrep -c .semgrep-service-name"${idx}".yml 2>&1 )
while [[ "${results}" == *Invalid_argument* ]]; do
  echo "${results}"
  results=$( semgrep -c .semgrep-service-name"${idx}".yml 2>&1 )
done
if [[ ! "${results}" == *" 0 findings"* ]]; then
  echo "${results}"
  return 1
fi
echo "${results}"
