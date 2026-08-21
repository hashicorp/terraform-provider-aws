---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_archive"
description: |-
  Manages an AWS SES Mail Manager Archive.
---

# Resource: aws_mailmanager_archive

Manages an AWS SES Mail Manager Archive.

## Example Usage

### Basic Usage

```terraform
resource "aws_mailmanager_archive" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the archive.

The following arguments are optional:

* `kms_key_arn` - (Optional, Forces new resource) ARN of the KMS key used to encrypt the archive.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `retention` - (Optional) Retention policy for the archive. See [`retention` Block](#retention-block).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `retention` Block

The following arguments are required:

* `retention_period` - (Required) Retention period for the archive. Valid values: `THREE_MONTHS`, `SIX_MONTHS`, `NINE_MONTHS`, `ONE_YEAR`, `EIGHTEEN_MONTHS`, `TWO_YEARS`, `THIRTY_MONTHS`, `THREE_YEARS`, `FOUR_YEARS`, `FIVE_YEARS`, `SIX_YEARS`, `SEVEN_YEARS`, `EIGHT_YEARS`, `NINE_YEARS`, `TEN_YEARS`, `PERMANENT`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `archive_state` - Current state of the archive. Always set to `ACTIVE` and will only be set to `PENDING_DELETION` when the archive is deleted.
* `arn` - ARN of the archive.
* `created_timestamp` - Timestamp of when the archive was created.
* `id` - Identifier of the archive.
* `last_updated_timestamp` - Timestamp of when the archive was updated.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_mailmanager_archive.example
  identity = {
    id = "archive-id-12345678"
  }
}

resource "aws_mailmanager_archive" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) Identifier of the archive.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an SES Mail Manager Archive using its identifier. For example:

```terraform
import {
  to = aws_mailmanager_archive.example
  id = "archive-id-12345678"
}

resource "aws_mailmanager_archive" "example" {
  ### Configuration omitted for brevity ###
}
```

Using `terraform import`, import an SES Mail Manager Archive using its identifier. For example:

```console
% terraform import aws_mailmanager_archive.example archive-id-12345678
```
