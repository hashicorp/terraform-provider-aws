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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `ids` - Set of Outposts Site identifiers.
