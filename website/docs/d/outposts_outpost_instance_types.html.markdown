---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outpost_instance_types"
description: |-
  Information about Outpost Instance Types.
---

# Data Source: aws_outposts_outpost_instance_types

Information about Outposts Instance Types.

## Example Usage

```terraform
data "aws_outposts_outpost_instance_types" "example" {
  arn = data.aws_outposts_outpost.example.arn
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `arn` - (Required) Outpost ARN.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `instance_types` - Set of instance types.
