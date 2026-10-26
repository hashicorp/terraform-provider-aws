# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_mailmanager_ingress_point" "test" {
  count  = var.resource_count
  region = var.region

  name              = "${var.rName}-${count.index}"
  type              = "OPEN"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id
}

resource "aws_mailmanager_traffic_policy" "test" {
  region = var.region

  default_action = "ALLOW"
  name           = var.rName

  policy_statement {
    action = "DENY"

    condition {
      ip_expression {
        operator = "CIDR_MATCHES"
        values   = ["192.0.2.0/24"]

        evaluate {
          attribute = "SENDER_IP"
        }
      }
    }
  }
}

resource "aws_mailmanager_rule_set" "test" {
  region = var.region

  name = var.rName

  rule {
    action {
      add_header {
        header_name  = "X-Test"
        header_value = "example"
      }
    }
  }
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
