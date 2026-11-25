---
subcategory: "Timestream for InfluxDB"
layout: "aws"
page_title: "AWS: aws_timestreaminfluxdb_db_backup"
description: |-
  Lists Timestream for InfluxDB DB Backup resources.
---

# List Resource: aws_timestreaminfluxdb_db_backup

Lists Timestream for InfluxDB DB Backup resources. Only on-demand backups are listed; automated backups are managed by the service and are not returned.

## Example Usage

### Basic Usage

```terraform
list "aws_timestreaminfluxdb_db_backup" "example" {
  provider = aws
}
```

### Filter by DB Resource

```terraform
list "aws_timestreaminfluxdb_db_backup" "example" {
  provider = aws

  config {
    db_resource_id = "abcdef1234"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `db_resource_id` - (Optional) ID of the DB instance or DB cluster to list backups for.
* `region` - (Optional) Region to query. Defaults to provider region.
