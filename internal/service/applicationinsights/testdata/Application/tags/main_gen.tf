# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name

  tags = var.resource_tags
}

resource "aws_resourcegroups_group" "test" {
  name = var.rName

  resource_query {
    query = <<JSON
	{
		"ResourceTypeFilters": [
		  "AWS::EC2::Instance"
		],
		"TagFilters": [
		  {
			"Key": "Stage",
			"Values": [
			  "Test"
			]
		  }
		]
	  }
JSON
  }
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
