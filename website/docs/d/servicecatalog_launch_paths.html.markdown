---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_launch_paths"
description: |-
  Provides information on Service Catalog Launch Paths
---

# Data Source: aws_servicecatalog_launch_paths

Lists the paths to the specified product. A path is how the user has access to a specified product, and is necessary when provisioning a product. A path also determines the constraints put on the product.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalog_launch_paths" "example" {
  product_id = "prod-yakog5pdriver"
}
```

## Argument Reference

The following arguments are required:

* `product_id` - (Required) Product identifier.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `summaries` - Block with information about the launch path. See details below.

### summaries

* `constraint_summaries` - Block for constraints on the portfolio-product relationship. See details below.
* `path_id` - Identifier of the product path.
* `name` - Name of the portfolio to which the path was assigned.
* `tags` - Tags associated with this product path.

### constraint_summaries

* `description` - Description of the constraint.
* `type` - Type of constraint. Valid values are `LAUNCH`, `NOTIFICATION`, `STACKSET`, and `TEMPLATE`.
