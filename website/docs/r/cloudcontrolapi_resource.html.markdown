---
subcategory: "Cloud Control API"
layout: "aws"
page_title: "AWS: aws_cloudcontrolapi_resource"
description: |-
    Manages a Cloud Control API Resource.
---

# Resource: aws_cloudcontrolapi_resource

Manages a Cloud Control API Resource. The configuration and lifecycle handling of these resources is proxied through Cloud Control API handlers to the backend service.

## Example Usage

```terraform
resource "aws_cloudcontrolapi_resource" "example" {
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

* `desired_state` - (Required) JSON string matching the CloudFormation resource type schema with desired configuration. Terraform configuration expressions can be converted into JSON using the [`jsonencode()` function](https://www.terraform.io/docs/language/functions/jsonencode.html).
* `type_name` - (Required) CloudFormation resource type name. For example, `AWS::EC2::VPC`.

The following arguments are optional:

* `role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role to assume for operations.
* `schema` - (Optional) JSON string of the CloudFormation resource type schema which is used for plan time validation where possible. Automatically fetched if not provided. In large scale environments with multiple resources using the same `type_name`, it is recommended to fetch the schema once via the [`aws_cloudformation_type` data source](/docs/providers/aws/d/cloudformation_type.html) and use this argument to reduce `DescribeType` API operation throttling. This value is marked sensitive only to prevent large plan differences from showing.
* `type_version_id` - (Optional) Identifier of the CloudFormation resource type version.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `properties` - JSON string matching the CloudFormation resource type schema with current configuration. Underlying attributes can be referenced via the [`jsondecode()` function](https://www.terraform.io/docs/language/functions/jsondecode.html), for example, `jsondecode(data.aws_cloudcontrolapi_resource.example.properties)["example"]`.
