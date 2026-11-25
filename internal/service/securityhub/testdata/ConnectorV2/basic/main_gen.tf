# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_connector_v2" "test" {

  name = var.rName

  connector_provider {
    jira_cloud {
      project_key = "TEST"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}

resource "aws_securityhub_aggregator_v2" "test" {
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = ["ap-southeast-1"]

  depends_on = [aws_securityhub_account_v2.test]
}

resource "aws_securityhub_account_v2" "test" {
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
