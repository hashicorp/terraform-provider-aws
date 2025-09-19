# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssmcontacts_contact" "test" {
  alias = "test-contact-for-${var.rName}"
  type  = "PERSONAL"

  depends_on = [data.aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact_channel" "test" {
  contact_id = aws_ssmcontacts_contact.test.arn

  delivery_address {
    simple_address = "test@example.com"
  }

  name = var.rName
  type = "EMAIL"
}

# testAccContactChannelConfig_base

data "aws_ssmincidents_replication_set" "test" {}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
