# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_batch_job_definition" "test" {
  provider = aws

  include_resource = true
}
