# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {
  region = var.region
}

resource "aws_ssoadmin_permission_set" "test" {
  region = var.region

  name         = var.rName
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachment" "test" {
  count = length(local.policy_names)

  region = var.region

  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/${local.policy_names[count.index]}"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}

locals {
  policy_names = ["AmazonCognitoReadOnly", "AmazonDynamoDBReadOnlyAccess"]
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
