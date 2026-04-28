---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_managed_resource_visibility"
description: |-
  Manages the default visibility of AWS-managed EC2 resources in the account and region.
---

# Resource: aws_ec2_managed_resource_visibility

Manages the default visibility of AWS-managed EC2 resources (such as managed prefix lists for AWS services) within an AWS account and region. Setting visibility to `hidden` excludes managed resources from default `Describe*` API responses and console listings.

~> **NOTE:** Only one `aws_ec2_managed_resource_visibility` resource may be configured per AWS account and region. Destroying this resource resets `default_visibility` to `visible`.

## Example Usage

### Hide AWS-managed resources

```terraform
resource "aws_ec2_managed_resource_visibility" "example" {
  default_visibility = "hidden"
}
```

### Restore default visibility

```terraform
resource "aws_ec2_managed_resource_visibility" "example" {
  default_visibility = "visible"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `default_visibility` - (Required) Default visibility for AWS-managed EC2 resources. Valid values are `hidden` and `visible`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS region in which the visibility setting applies.
