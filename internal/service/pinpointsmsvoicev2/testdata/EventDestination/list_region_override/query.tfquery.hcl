# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_pinpointsmsvoicev2_event_destination" "test" {
  provider = aws

  config {
    configuration_set_names = [aws_pinpointsmsvoicev2_configuration_set.test.name]
    region                  = var.region
  }
}
