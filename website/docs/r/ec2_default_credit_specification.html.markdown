---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_default_credit_specification"
description: |-
  Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Default Credit Specification.
---
# Resource: aws_ec2_default_credit_specification

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Default Credit Specification.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_default_credit_specification" "example" {
  instance_family = "t2"
  cpu_credits     = "standard"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cpu_credits` - (Required) Credit option for CPU usage of the instance family. Valid values: `standard`, `unlimited`.
* `instance_family` - (Required) Instance family. Valid values are `t2`, `t3`, `t3a`, `t4g`.

## Attribute Reference

This data source exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 (Elastic Compute Cloud) Default Credit Specification using the `instance_family`. For example:

```terraform
import {
  to = aws_ec2_default_credit_specification.example
  id = "t2"
}
```

Using `terraform import`, import EC2 (Elastic Compute Cloud) Default Credit Specification using the `instance_family`. For example:

```console
% terraform import aws_ec2_default_credit_specification.example t2
