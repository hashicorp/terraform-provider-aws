# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_account_default_protect_configuration" "test" {
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.test.id
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
}

