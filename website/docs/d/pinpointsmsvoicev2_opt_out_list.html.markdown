---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_opt_out_list"
description: |-
  Provides details about an AWS End User Messaging SMS opt-out list.
---

# Data Source: aws_pinpointsmsvoicev2_opt_out_list

Provides details about an AWS End User Messaging SMS opt-out list.

## Example Usage

### Basic Usage

```terraform
data "aws_pinpointsmsvoicev2_opt_out_list" "example" {
  name = "example-opt-out-list"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the opt-out list.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the opt-out list.
* `tags` - Map of tags assigned to the opt-out list.
