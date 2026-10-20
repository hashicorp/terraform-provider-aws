# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

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

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
