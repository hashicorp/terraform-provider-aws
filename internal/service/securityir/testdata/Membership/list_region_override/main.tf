# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityir_membership" "test" {
  region = var.region

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
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
