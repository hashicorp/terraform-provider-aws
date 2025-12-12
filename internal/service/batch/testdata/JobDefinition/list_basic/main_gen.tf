# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_batch_job_definition" "test" {
  count = 2

  name = "${var.rName}-${count.index}"
  type = "container"
  container_properties = jsonencode({
    image  = "busybox"
    vcpus  = 1
    memory = 128
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
