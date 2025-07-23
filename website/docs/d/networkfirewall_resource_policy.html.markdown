---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_resource_policy"
description: |-
  Retrieve information about a Network Firewall resource policy.
---

# Data Source: aws_networkfirewall_resource_policy

Retrieve information about a Network Firewall resource policy.

## Example Usage

```terraform
data "aws_networkfirewall_resource_policy" "example" {
  resource_arn = var.resource_policy_arn
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_arn` - (Required) The Amazon Resource Name (ARN) that identifies the resource policy.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) that identifies the resource policy.
* `policy` - The [policy][1] for the resource.

[1]: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/networkfirewall_resource_policy
