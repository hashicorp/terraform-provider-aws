---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outpost"
description: |-
  Provides details about an Outposts Outpost
---

# Data Source: aws_outposts_outpost

Provides details about an Outposts Outpost.

## Example Usage

```terraform
data "aws_outposts_outpost" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Optional) ARN.
* `id` - (Optional) Identifier of the Outpost.
* `name` - (Optional) Name of the Outpost.
* `owner_id` - (Optional) AWS Account identifier of the Outpost owner.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `availability_zone` - Availability Zone name.
* `availability_zone_id` - Availability Zone identifier.
* `description` - Description of the Outpost.
* `lifecycle_status` - Life cycle status.
* `site_arn` - ARN of the site.
* `site_id` - ID of the site.
* `supported_hardware_type` - Hardware type.
* `tags` - Outpost tags.
