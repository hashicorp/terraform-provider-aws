#!/bin/bash

idx=$1

results=$( semgrep -c .semgrep-service-name"${idx}".yml 2>&1 )
while [[ "${results}" == *Invalid_argument* ]]; do
  echo "${results}"
  results=$( semgrep -c .semgrep-service-name"${idx}".yml 2>&1 )
done
if [[ ! "${results}" == *" 0 findings"* ]]; then
  >&2 echo "${results}"
fi
echo "${results}"
