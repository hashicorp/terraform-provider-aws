---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_event_categories"
description: |-
    Provides a list of DB Event Categories which can be used to pass values into DB Event Subscription.
---

# Data Source: aws_db_event_categories

## Example Usage

List the event categories of all the RDS resources.

```terraform
data "aws_db_event_categories" "example" {}

output "example" {
  value = data.aws_db_event_categories.example.event_categories
}
```

List the event categories specific to the RDS resource `db-snapshot`.

```terraform
data "aws_db_event_categories" "example" {
  source_type = "db-snapshot"
}

output "example" {
  value = data.aws_db_event_categories.example.event_categories
}
```

## Argument Reference

This data source supports the following arguments:

* `source_type` - (Optional) Type of source that will be generating the events. Valid options are db-instance, db-security-group, db-parameter-group, db-snapshot, db-cluster or db-cluster-snapshot.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `event_categories` - List of the event categories.
* `id` - Region of the event categories.
