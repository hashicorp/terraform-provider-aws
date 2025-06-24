---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_endpoint"
description: |-
  Terraform data source for managing an AWS DMS (Database Migration) Endpoint.
---

# Data Source: aws_dms_endpoint

Terraform data source for managing an AWS DMS (Database Migration) Endpoint.

## Example Usage

### Basic Usage

```terraform
data "aws_dms_endpoint" "test" {
  endpoint_id = "test_id"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `endpoint_id` - (Required) Database endpoint identifier. Identifiers must contain from 1 to 255 alphanumeric characters or hyphens, begin with a letter, contain only ASCII letters, digits, and hyphens, not end with a hyphen, and not contain two consecutive hyphens.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

See the [`aws_dms_endpoint` resource](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/dms_endpoint) for details on the returned attributes - they are identical.
