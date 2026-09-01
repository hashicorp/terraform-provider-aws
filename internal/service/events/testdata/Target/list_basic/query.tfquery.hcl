# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_cloudwatch_event_target" "test" {
  provider = aws
  config {
    event_bus_name = "default"
    rule           = var.rName
  }
}
