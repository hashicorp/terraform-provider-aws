---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_ip_restriction"
description: |-
  Manages the content and status of IP rules.
---

# Resource: aws_quicksight_ip_restriction

Manages the content and status of IP rules.

~> Deletion of this resource clears all IP restrictions from a QuickSight account.

## Example Usage

```terraform
resource "aws_quicksight_ip_restriction" "example" {
  enabled = true

  ip_restriction_rule_map = {
    "108.56.166.202/32" = "Allow self"
  }

  vpc_id_restriction_rule_map = {
    (aws_vpc.example.id) = "Main VPC"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `enabled` - (Required) Whether IP rules are turned on.
* `ip_restriction_rule_map` - (Optional) Map of allowed IPv4 CIDR ranges and descriptions.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_endpoint_id_restriction_rule_map` - (Optional) Map of allowed VPC endpoint IDs and descriptions.
* `vpc_id_restriction_rule_map` - (Optional) Map of VPC IDs and descriptions. Traffic from all VPC endpoints that are present in the specified VPC is allowed.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight IP restriction using the AWS account ID. For example:

```terraform
import {
  to = aws_quicksight_ip_restriction.example
  id = "012345678901"
}
```

Using `terraform import`, import QuickSight IP restriction using the AWS account ID. For example:

```console
% terraform import aws_quicksight_ip_restriction.example "012345678901"
```
