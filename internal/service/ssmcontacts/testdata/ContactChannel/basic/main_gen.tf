# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssmcontacts_contact_channel" "test" {
  contact_id = aws_ssmcontacts_contact.test.arn

  delivery_address {
    simple_address = "test@example.com"
  }

  name = var.rName
  type = "EMAIL"
}

resource "aws_ssmcontacts_contact" "test" {
  alias = "test-contact-for-${var.rName}"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

# testAccContactChannelConfig_base

resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = data.aws_region.current.region
  }
}

data "aws_region" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
