#!/bin/bash

idx=$1

results=$( semgrep -c .semgrep-service-name"${idx}".yml 2>&1 )
while [[ "${results}" == *Invalid_argument* ]] && [[ "${results}" == *" 0 findings"* ]]; do
  echo "${results}"
  results=$( semgrep -c .semgrep-service-name"${idx}".yml 2>&1 )
done
if [[ ! "${results}" == *" 0 findings"* ]]; then
  echo "${results}" >&2
  exit 1
fi
echo "${results}"
