# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_batch_job_definition" "test" {
  count = var.resource_count

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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
