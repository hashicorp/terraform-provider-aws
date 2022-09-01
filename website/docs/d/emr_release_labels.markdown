---
subcategory: "EMR"
layout: "aws"
page_title: "AWS: aws_emr_release_labels"
description: |-
  Retrieve information about EMR Release Labels
---

# Data Source: aws_emr_release_labels

Retrieve information about EMR Release Labels.

## Example Usage

```terraform
data "aws_emr_release_labels" "example" {
  filters {
    application = "spark@2.1.0"
    prefix      = "emr-5"
  }
}
```

## Argument Reference

* `filters` – (Optional) Filters the results of the request. Prefix specifies the prefix of release labels to return. Application specifies the application (with/without version) of release labels to return. See [Filters](#filters).

### Filters

* `application` - (Optional) Optional release label application filter. For example, `Spark@2.1.0` or `Spark`.
* `prefix` - (Optional) Optional release label version prefix filter. For example, `emr-5`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `release_labels` - Returned release labels.
