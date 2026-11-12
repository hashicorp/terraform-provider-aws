# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_securityir_membership" "test" {
  membership_name = var.rName

  incident_response_team {
    name      = "Responder One"
    job_title = "Security Engineer"
    email     = "${var.rName}-1@example.com"
  }

  incident_response_team {
    name      = "Responder Two"
    job_title = "Security Manager"
    email     = "${var.rName}-2@example.com"
  }

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = false
}
