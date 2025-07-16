# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = "tf-acc-test-custom-routing-accelerator"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
