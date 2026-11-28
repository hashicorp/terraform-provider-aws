---
subcategory: "Bedrock Data Automation"
layout: "aws"
page_title: "AWS: aws_bedrockdataautomation_project"
description: |-
  Lists Amazon Bedrock Data Automation Project resources.
---

# List Resource: aws_bedrockdataautomation_project

Lists Amazon Bedrock Data Automation Project resources. Only projects owned by the account are listed. With `include_resource`, `standard_output_configuration` and `override_configuration` aren't populated.

## Example Usage

```terraform
list "aws_bedrockdataautomation_project" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
