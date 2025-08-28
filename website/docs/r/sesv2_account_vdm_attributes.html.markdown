---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_account_vdm_attributes"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Account VDM Attributes.
---

# Resource: aws_sesv2_account_vdm_attributes

Terraform resource for managing an AWS SESv2 (Simple Email V2) Account VDM Attributes.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_account_vdm_attributes" "example" {
  vdm_enabled = "ENABLED"

  dashboard_attributes {
    engagement_metrics = "ENABLED"
  }

  guardian_attributes {
    optimized_shared_delivery = "ENABLED"
  }
}
```

## Argument Reference

The following arguments are required:

* `vdm_enabled` - (Required) Specifies the status of your VDM configuration. Valid values: `ENABLED`, `DISABLED`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `dashboard_attributes` - (Optional) Specifies additional settings for your VDM configuration as applicable to the Dashboard.
* `guardian_attributes` - (Optional) Specifies additional settings for your VDM configuration as applicable to the Guardian.

### dashboard_attributes

* `engagement_metrics` - (Optional) Specifies the status of your VDM engagement metrics collection. Valid values: `ENABLED`, `DISABLED`.

### guardian_attributes

* `optimized_shared_delivery` - (Optional) Specifies the status of your VDM optimized shared delivery. Valid values: `ENABLED`, `DISABLED`.

## Attribute Reference

This data source exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Account VDM Attributes using the word `ses-account-vdm-attributes`. For example:

```terraform
import {
  to = aws_sesv2_account_vdm_attributes.example
  id = "ses-account-vdm-attributes"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Account VDM Attributes using the word `ses-account-vdm-attributes`. For example:

```console
% terraform import aws_sesv2_account_vdm_attributes.example ses-account-vdm-attributes
```
