# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_bedrockagentcore_registry_record" "test" {
  provider = aws

  config {
    region      = var.region
    registry_id = aws_bedrockagentcore_registry.test.registry_id
  }
}
