---
subcategory: "Savings Plans"
layout: "aws"
page_title: "AWS: aws_savingsplans_offerings"
description: |-
  Data source for getting AWS Savings Plans Offerings.
---

# Data Source: aws_savingsplans_offerings

Data source for getting AWS Savings Plans Offerings.

## Example Usage

### Basic Usage

```terraform
data "aws_savingsplans_offerings" "example" {
  product_type = "EC2"

  filter {
    name   = "region"
    values = ["us-west-2"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `currencies` - (Optional) List of currencies.
* `descriptions` - (Optional) List of descriptions.
* `durations` - (Optional) List of durations, in seconds.
* `filter` - (Optional) List of filters. See [Filter](#filter).
* `offering_ids` - (Optional) List of offering IDs.
* `operations` - (Optional) List of operations.
* `payment_options` - (Optional) List of payment options.
* `plan_types` - (Optional) List of plan types.
* `product_type` - (Optional) Product type.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_codes` - (Optional) List of service codes.
* `usage_types` - (Optional) List of usage types.

### Filter

* `name` - (Required) Filter name.
* `values` - (Required) List of filter values.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `offerings` - List of Savings Plans Offerings. See [`offerings` Attribute Reference](#offerings-attribute-reference).

### `offerings` Attribute Reference

* `currency` - Currency.
* `description` - Description.
* `duration_seconds` - Duration, in seconds.
* `offering_id` - Offering ID.
* `operation` - Operation.
* `payment_option` - Payment option.
* `plan_type` - Plan type.
* `product_types` - List of product types.
* `properties` - List of properties. See [`properties` Attribute Reference](#properties-attribute-reference).
* `service_code` - Service code.
* `usage_type` - Usage type.

### `properties` Attribute Reference

* `name` - Property name.
* `value` - Property value.
