# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_key_pair" "test" {
  count  = var.resource_count
  region = var.region

  key_name   = "${var.rName}-${count.index}"
  public_key = count.index == 0 ? var.public_key_1 : var.public_key_2
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

variable "public_key_1" {
  description = "Public key material for resource 0"
  type        = string
  nullable    = false
}

variable "public_key_2" {
  description = "Public key material for resource 1"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
