---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_domain"
description: |-
  Terraform data source for managing an AWS DataZone Domain.
---

# Data Source: aws_datazone_domain

Terraform data source for managing an AWS DataZone Domain.

## Example Usage

### Basic Usage

```terraform
data "aws_datazone_domain" "example" {
  name = "example_domain"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Optional) Name of the Domain. One of `name` or `id` is required.
* `id` - (Optional) ID of the Domain. One of `name` or `id` is required

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Domain.
* `created_at` - The date and time the Domain was created.
* `description` - Description of the Domain.
* `domain_version` - Version of the Domain.
* `last_updated_at` - The date and time the Domain was last updated.
* `managed_account_id` - The AWS account ID that owns the Domain.
* `portal_url` - URL of the Domain.
* `status` - Status of the Domain.
