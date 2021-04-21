---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_resource"
description: |-
    Provides details for a CloudFormation Resource.
---

# Data Source: aws_cloudformation_resource

Provides details for a CloudFormation Resource. The reading of these resources is proxied through CloudFormation handlers to the backend service.

## Example Usage

```terraform
resource "aws_cloudformation_resource" "example" {
  identifier = "example"
  type_name  = "AWS::ECS::Cluster"
}
```

## Argument Reference

The following arguments are required:

* `identifier` - (Required) "Identifier of the CloudFormation resource type. For example, `vpc-12345678`."
* `type_name` - (Required) CloudFormation resource type name. For example, `AWS::EC2::VPC`.

The following arguments are optional:

* `role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role to assume for operations.
* `type_version_id` - (Optional) Identifier of the CloudFormation resource type version.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `resource_model` - JSON matching the CloudFormation resource type schema with current configuration.
