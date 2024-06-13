# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
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
