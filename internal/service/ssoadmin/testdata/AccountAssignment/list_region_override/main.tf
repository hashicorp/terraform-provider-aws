# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_ssoadmin_instances" "test" {
  region = var.region
}

data "aws_caller_identity" "current" {}

resource "aws_ssoadmin_permission_set" "test" {
  region = var.region

  name         = var.rName
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

data "aws_identitystore_group" "group" {
  region = var.region

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "DisplayName"
      attribute_value = var.AWS_IDENTITY_STORE_GROUP_NAME
    }
  }
}

data "aws_identitystore_user" "user" {
  region = var.region

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "UserName"
      attribute_value = var.AWS_IDENTITY_STORE_USER_NAME
    }
  }
}

resource "aws_ssoadmin_account_assignment" "group" {
  region = var.region

  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "GROUP"
  principal_id       = data.aws_identitystore_group.group.group_id
}

resource "aws_ssoadmin_account_assignment" "user" {
  region = var.region

  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "USER"
  principal_id       = data.aws_identitystore_user.user.user_id
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "AWS_IDENTITY_STORE_GROUP_NAME" {
  type     = string
  nullable = false
}

variable "AWS_IDENTITY_STORE_USER_NAME" {
  type     = string
  nullable = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
