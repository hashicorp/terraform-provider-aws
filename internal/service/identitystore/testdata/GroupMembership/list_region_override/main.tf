# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_ssoadmin_instances" "test" {
  region = var.region
}

resource "aws_identitystore_user" "test" {
  count  = var.resource_count
  region = var.region

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = "${var.rName}-${count.index}"
  user_name         = "${var.rName}-${count.index}"

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}

resource "aws_identitystore_group" "test" {
  region = var.region

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = var.rName
  description       = "Acceptance Test"
}

resource "aws_identitystore_group_membership" "test" {
  count  = var.resource_count
  region = var.region

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  group_id          = aws_identitystore_group.test.group_id
  member_id         = aws_identitystore_user.test[count.index].user_id
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
