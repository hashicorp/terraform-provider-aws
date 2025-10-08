---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_site"
description: |-
  Provides details about an Outposts Site
---

# Data Source: aws_outposts_site

Provides details about an Outposts Site.

## Example Usage

```terraform
data "aws_outposts_site" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `id` - (Optional) Identifier of the Site.
* `name` - (Optional) Name of the Site.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `account_id` - AWS Account identifier.
* `description` - Description.
