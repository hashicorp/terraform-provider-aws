---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_instance_profile"
description: |-
  Lists DMS (Database Migration) Instance Profile resources.
---

# List Resource: aws_dms_instance_profile

Lists DMS (Database Migration) Instance Profile resources.

## Example Usage

```terraform
list "aws_dms_instance_profile" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
