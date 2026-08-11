# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_caller_identity" "current" {}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = var.rName
  db_name        = "test"
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = var.rName
}

resource "aws_redshift_namespace_registration" "test" {
  consumer_identifier             = format("DataCatalog/%s", data.aws_caller_identity.current.account_id)
  namespace_type                  = "serverless"
  serverless_namespace_identifier = aws_redshiftserverless_namespace.test.namespace_name
  serverless_workgroup_identifier = aws_redshiftserverless_workgroup.test.workgroup_name
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
