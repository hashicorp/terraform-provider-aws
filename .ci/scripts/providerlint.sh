#!/bin/bash

results=$( make providerlint 2>&1 )
while [[ "${results}" == *Killed* ]]; do
  echo "${results}"
  results=$( make providerlint 2>&1 )
done
echo "${results}"

rules=(
    # Syntax checks
    "--only=terraform_comment_syntax"
    "--only=terraform_deprecated_index"
    "--only=terraform_deprecated_interpolation"
    # Ensure valid instance types
    "--only=aws_db_instance_invalid_type"
    # Ensure modern instance types
    "--only=aws_db_instance_previous_type"
    "--only=aws_instance_previous_type"
    # Ensure engine types are valid
    "--only=aws_db_instance_invalid_engine"
    "--only=aws_mq_broker_invalid_engine_type"
)
