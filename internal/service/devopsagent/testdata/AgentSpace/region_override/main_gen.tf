# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_devopsagent_agent_space" "test" {
  region = var.region

  name = "tf-acc-test-devopsagent"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
