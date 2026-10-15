---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_templates"
description: |-
  Terraform data source for managing AWS Service Quotas Templates.
---

# Data Source: aws_servicequotas_templates

Terraform data source for managing AWS Service Quotas Templates.

## Example Usage

### Basic Usage

```terraform
data "aws_servicequotas_templates" "example" {
  aws_region = "us-east-1"
}
```

## Argument Reference

This data source supports the following arguments:

* `aws_region` - (Optional) AWS Region to which the quota increases apply.
* `region` - (Optional, **Deprecated**) AWS Region to which the quota increases apply. Use `aws_region` instead.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `templates` - A list of quota increase templates for specified region. See [`templates`](#templates).

### `templates`

* `global_quota` - Indicates whether the quota is global.
* `quota_name` - Quota name.
* `quota_code` - Quota identifier.
* `region` - AWS Region to which the template applies.
* `service_code` - Service identifier.
* `service_name` - Service name.
* `unit` - Unit of measurement.
* `value` - The new, increased value for the quota.
