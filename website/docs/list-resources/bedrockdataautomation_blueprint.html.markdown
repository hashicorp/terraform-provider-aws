---
subcategory: "Bedrock Data Automation"
layout: "aws"
page_title: "AWS: aws_bedrockdataautomation_blueprint"
description: |-
  Lists Amazon Bedrock Data Automation Blueprint resources.
---

# List Resource: aws_bedrockdataautomation_blueprint

Lists Amazon Bedrock Data Automation Blueprint resources. Only blueprints owned by the account are listed.

## Example Usage

```terraform
list "aws_bedrockdataautomation_blueprint" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
