# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_batch_job_definition" "test" {
  name = var.rName
  type = "container"
  container_properties = jsonencode({
    command = ["echo", "test"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })

  tags = var.tags
}

variable "rName" {
  type     = string
  nullable = false
}

variable "tags" {
  type     = map(string)
  nullable = false
}


