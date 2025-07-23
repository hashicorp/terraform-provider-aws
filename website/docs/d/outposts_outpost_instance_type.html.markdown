---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outpost_instance_type"
description: |-
  Information about single Outpost Instance Type.
---

# Data Source: aws_outposts_outpost_instance_type

Information about single Outpost Instance Type.

## Example Usage

```terraform
data "aws_outposts_outpost_instance_type" "example" {
  arn                      = data.aws_outposts_outpost.example.arn
  preferred_instance_types = ["m5.large", "m5.4xlarge"]
}

resource "aws_ec2_instance" "example" {
  # ... other configuration ...

  instance_type = data.aws_outposts_outpost_instance_type.example.instance_type
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) Outpost ARN.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `instance_type` - (Optional) Desired instance type. Conflicts with `preferred_instance_types`.
* `preferred_instance_types` - (Optional) Ordered list of preferred instance types. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned. Conflicts with `instance_type`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Outpost identifier.
