---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_opt_out_list"
description: |-
  Manages an AWS End User Messaging SMS opt-out list.
---

# Resource: aws_pinpointsmsvoicev2_opt_out_list

Manages an AWS End User Messaging SMS opt-out list.

## Example Usage

```terraform
resource "aws_pinpointsmsvoicev2_opt_out_list" "example" {
  name = "example-opt-out-list"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the opt-out list.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the opt-out list.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import opt-out lists using the `name`. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_opt_out_list.example
  id = "example-opt-out-list"
}
```

Using `terraform import`, import opt-out lists using the `name`. For example:

```console
% terraform import aws_pinpointsmsvoicev2_opt_out_list.example example-opt-out-list
```
