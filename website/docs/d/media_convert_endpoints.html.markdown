---
subcategory: "Elemental MediaConvert"
layout: "aws"
page_title: "AWS: aws_media_convert_endpoints"
description: |-
    Provides details about all MediaConvert endpoints.
---

# Data Source: aws_media_convert_endpoints

Provides details about all MediaConvert endpoints.

## Example Usage

### All endpoints

```terraform
data "aws_media_convert_endpoints" "all" {
}
```

## Argument Reference

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `endpoints` - List with map of strings to urls.
