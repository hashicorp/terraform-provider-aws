---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_resource"
description: |-
    Manages a CloudFormation Resource.
---

# Resource: aws_cloudformation_resource

Manages a CloudFormation Resource. The configuration and lifecycle handling of these resources is proxied through CloudFormation handlers to the backend service.

## Example Usage

```terraform
resource "aws_cloudformation_resource" "example" {
  type_name = "AWS::ECS::Cluster"

  desired_state = jsonencode({
    ClusterName = "example"
    Tags = [
      {
        Key   = "CostCenter"
        Value = "IT"
      }
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `desired_state` - (Required) JSON matching the CloudFormation resource type schema with desired configuration.
* `type_name` - (Required) CloudFormation resource type name. For example, `AWS::EC2::VPC`.
* `type_version_id` - (Required) Identifier of the CloudFormation resource type version.

The following arguments are optional:

* `role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role to assume for operations.
* `schema` - (Optional) JSON of the CloudFormation resource type schema. Automatically fetched if not provided. This value is marked sensitive only to prevent large plan differences from showing.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `resource_model` - JSON matching the CloudFormation resource type schema with current configuration.
