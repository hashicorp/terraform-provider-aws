---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_sites"
description: |-
  Provides details about multiple Outposts Sites.
---

# Data Source: aws_outposts_sites

Provides details about multiple Outposts Sites.

## Example Usage

```terraform
data "aws_outposts_sites" "all" {}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `ids` - Set of Outposts Site identifiers.
