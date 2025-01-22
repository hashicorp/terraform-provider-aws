# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_servicecatalogappregistry_attribute_group" "test" {
  id = aws_servicecatalogappregistry_attribute_group.test.id
}
resource "aws_servicecatalogappregistry_attribute_group" "test" {
  name        = var.rName
  description = "Some attribute group"

  attributes = jsonencode({
    a = "1"
    b = "2"
  })

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
