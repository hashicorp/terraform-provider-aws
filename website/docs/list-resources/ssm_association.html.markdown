---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_association"
description: |-
  Lists SSM (Systems Manager) Association resources.
---

# List Resource: aws_ssm_association

Lists SSM (Systems Manager) Association resources..

## Example Usage

```terraform
list "aws_ssm_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query.
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
