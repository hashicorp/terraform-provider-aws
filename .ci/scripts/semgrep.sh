#!/bin/bash

cfile=$1

results=$( semgrep -c "${cfile}" 2>&1 )
while [[ "${results}" == *Invalid_argument* ]] && [[ "${results}" == *" 0 findings"* ]]; do
  echo "${results}"
  results=$( semgrep -c "${cfile}" 2>&1 )
done
if [[ ! "${results}" == *" 0 findings"* ]]; then
  echo "${results}" >&2
  exit 1
fi
echo "${results}"
