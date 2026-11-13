# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_appautoscaling_policy" "test" {
  provider = aws

  include_resource = true

  config {
    service_namespace = "ecs"
  }
}
