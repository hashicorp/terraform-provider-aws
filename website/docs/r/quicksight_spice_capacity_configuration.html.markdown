---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_spice_capacity_configuration"
description: |-
  Manages the SPICE capacity purchase mode for a QuickSight account.
---

# Resource: aws_quicksight_spice_capacity_configuration

Manages the [SPICE](https://docs.aws.amazon.com/quicksight/latest/user/spice.html) capacity purchase mode for a QuickSight account. This controls whether extra SPICE capacity is purchased automatically as needed, or only manually.

~> **Note:** QuickSight does not provide an API to read the current SPICE purchase mode. As a result, this resource is write-only: the configured `purchase_mode` is stored in Terraform state but cannot be refreshed, and changes made outside of Terraform (drift) cannot be detected.

~> **Warning:** Destroying this resource resets the account's SPICE purchase mode to `MANUAL` (the QuickSight default), disabling automatic SPICE capacity purchasing.

## Example Usage

### Automatic SPICE capacity purchasing

```terraform
resource "aws_quicksight_spice_capacity_configuration" "example" {
  purchase_mode = "AUTO_PURCHASE"
}
```

### Manual SPICE capacity purchasing

```terraform
resource "aws_quicksight_spice_capacity_configuration" "example" {
  purchase_mode = "MANUAL"
}
```

## Argument Reference

This resource supports the following arguments:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `purchase_mode` - (Optional) Determines how SPICE capacity can be purchased. Valid values are `MANUAL` (SPICE capacity can only be purchased manually) and `AUTO_PURCHASE` (extra SPICE capacity is automatically purchased on your behalf as needed; SPICE capacity can also still be purchased manually). Defaults to `MANUAL`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

~> **Note:** Because the SPICE purchase mode cannot be read from the QuickSight API, `purchase_mode` is not populated on import and will show a difference on the next plan.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight SPICE capacity configuration using the AWS account ID. For example:

```terraform
import {
  to = aws_quicksight_spice_capacity_configuration.example
  id = "012345678901"
}
```

Using `terraform import`, import QuickSight SPICE capacity configuration using the AWS account ID. For example:

```console
% terraform import aws_quicksight_spice_capacity_configuration.example "012345678901"
```
