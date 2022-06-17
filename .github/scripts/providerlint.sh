#!/bin/bash

idx=$1

results=$( make providerlint 2>&1 )
while [[ "${results}" == *Killed* ]]; do
  echo "${results}"
  results=$( make providerlint 2>&1 )
done
echo "${results}"
