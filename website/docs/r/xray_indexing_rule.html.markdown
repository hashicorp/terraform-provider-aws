---
subcategory: "X-Ray"
layout: "aws"
page_title: "AWS: aws_xray_indexing_rule"
description: |-
    Manages an AWS X-Ray indexing rule.
---

# Resource: aws_xray_indexing_rule

Manages an AWS X-Ray indexing rule.

-> **Note:** Removing this resource from Terraform has no effect on the indedxing rule within AWS X-Ray.

## Example Usage

```terraform
resource "aws_xray_indexing_rule" "example" {
  name = "Default"

  rule {
    probabilistic {
      desired_sampling_percentage = 0.66
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Indexing rule name.
* `rule` - (Required) Rule configuration. See [`rule` Block](#rule-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `rule` Block

The `rule` block supports:

* `probabilistic` - (Optional) Indexing rule configuration used to probabilistically sample traceIds. See [`probabilistic` Block](#probabilistic-block) below.

### `probabilistic` Block

The `probabilistic` block supports:

* `desired_sampling_percentage` - (Required) Configured sampling percentage of traceIds.
* `actual_sampling_percentage` - (Computed) Applied sampling percentage of traceIds.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_xray_indexing_rule.example
  identity = {
    name = "Default"
  }
}

resource "aws_xray_indexing_rule" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Indexing rule name.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import XRay Indexing Rules using `name`. For example:

```terraform
import {
  to = aws_xray_indexing_rule.example
  id = "Default"
}
```

Using `terraform import`, import XRay Indexing Rules using `name`. For example:

```console
% terraform import aws_xray_indexing_rule.example Default
```
