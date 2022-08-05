---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_key_pair"
description: |-
    Provides details about a specific EC2 Key Pair.
---

# Data Source: aws_key_pair

Use this data source to get information about a specific EC2 Key Pair.

## Example Usage

The following example shows how to get a EC2 Key Pair including the public key material from its name.

```terraform
data "aws_key_pair" "example" {
  key_name           = "test"
  include_public_key = true

  filter {
    name   = "tag:Component"
    values = ["web"]
  }
}

output "fingerprint" {
  value = data.aws_key_pair.example.fingerprint
}

output "name" {
  value = data.aws_key_pair.example.key_name
}

output "id" {
  value = data.aws_key_pair.example.id
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Key Pairs. The given filters must match exactly one Key Pair
whose data will be exported as attributes.

* `key_pair_id` - (Optional) The Key Pair ID.
* `key_name` - (Optional) The Key Pair name.
* `include_public_key` - (Optional) Whether to include the public key material in the response.
* `filter` -  (Optional) Custom filter block as described below.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the [EC2 DescribeKeyPairs API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeKeyPairs.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the Key Pair.
* `arn` - The ARN of the Key Pair.
* `create_time` - The timestamp for when the key pair was created in ISO 8601 format.
* `fingerprint` - The SHA-1 digest of the DER encoded private key.
* `key_type` - The type of key pair.
* `public_key` - The public key material.
* `tags` - Any tags assigned to the Key Pair.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

- `read` - (Default `20m`)
