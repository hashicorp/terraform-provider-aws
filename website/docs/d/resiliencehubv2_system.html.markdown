---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_system"
description: |-
  Terraform data source for reading an AWS Resilience Hub V2 System.
---

# Data Source: aws_resiliencehubv2_system

Terraform data source for reading an AWS Resilience Hub V2 System.

## Example Usage

### Basic Usage

```hcl
data "aws_resiliencehubv2_system" "example" {
  arn = "arn:aws:resiliencehub:us-west-2:123456789012:system/example-system:abc123"
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN of the system.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the system.
* `name` - Name of the system.
* `sharing_enabled` - Whether cross-account sharing is enabled.
* `tags` - Map of tags assigned to the resource.
