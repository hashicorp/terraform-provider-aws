# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_config_remediation_configuration" "test" {
  provider = aws

  include_resource = true

  config {
    config_rule_names = aws_config_config_rule.test.*.name
  }
}
