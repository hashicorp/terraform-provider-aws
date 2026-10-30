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

* `region` - (Optional) Region to query. Defaults to provider region.
