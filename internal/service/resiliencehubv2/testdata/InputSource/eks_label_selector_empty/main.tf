# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# An empty label_selector is rejected at plan time: the service drops a selector with neither
# matchLabels nor matchExpressions, so it would never read back. Nothing here is applied.
resource "aws_resiliencehubv2_input_source" "test" {
  service_arn = "arn:${data.aws_partition.current.partition}:resiliencehub:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:service/${var.rName}:abc123"

  resource_configuration {
    eks {
      cluster_arn = "arn:${data.aws_partition.current.partition}:eks:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:cluster/${var.rName}"
      namespaces  = ["default"]

      label_selector {
      }
    }
  }
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
