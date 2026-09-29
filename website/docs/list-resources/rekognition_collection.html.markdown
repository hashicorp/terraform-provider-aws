---
subcategory: "Rekognition"
layout: "aws"
page_title: "AWS: aws_rekognition_collection"
description: |-
  Lists Rekognition Collection resources.
---

# List Resource: aws_rekognition_collection

Lists Rekognition Collection resources.

## Example Usage

```terraform
list "aws_rekognition_collection" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
