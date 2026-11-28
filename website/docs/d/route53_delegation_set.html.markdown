---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_delegation_set"
description: |-
    Provides details about a specific Route 53 Delegation Set
---

# Data Source: aws_route53_delegation_set

`aws_route53_delegation_set` provides details about a specific Route 53 Delegation Set.

This data source allows to find a list of name servers associated with a specific delegation set.

## Example Usage

The following example shows how to get a delegation set from its ID.

```terraform
data "aws_route53_delegation_set" "dset" {
  id = "MQWGHCBFAKEID"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `id` - (Required) Delegation set ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Delegation Set.
* `caller_reference` - Caller Reference of the delegation set.
* `name_servers` - List of DNS name servers for the delegation set.
