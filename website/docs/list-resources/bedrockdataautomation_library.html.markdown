---
subcategory: "Bedrock Data Automation"
layout: "aws"
page_title: "AWS: aws_bedrockdataautomation_library"
description: |-
  Lists Amazon Bedrock Data Automation Library resources.
---

# List Resource: aws_bedrockdataautomation_library

Lists Amazon Bedrock Data Automation Library resources.

## Example Usage

```terraform
list "aws_bedrockdataautomation_library" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
