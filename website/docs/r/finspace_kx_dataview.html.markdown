---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_dataview"
description: |-
  Terraform resource for managing an AWS FinSpace Kx Dataview.
---

# Resource: aws_finspace_kx_dataview

Terraform resource for managing an AWS FinSpace Kx Dataview.

## Example Usage

### Basic Usage

```terraform
resource "aws_finspace_kx_dataview" "example" {
  name                 = "my-tf-kx-dataview"
  environment_id       = aws_finspace_kx_environment.example.id
  database_name        = aws_finspace_kx_database.example.name
  availability_zone_id = "use1-az2"
  description          = "Terraform managed Kx Dataview"
  az_mode              = "SINGLE"
  auto_update          = true

  segment_configurations {
    volume_name = aws_finspace_kx_volume.example.name
    db_paths    = ["/*"]
  }

  # Depending on the type of cache and size of the Kx Volume, create/update timeouts 
  # may need to be increased up to a potential maximum of 24 hours and the delete timeout to 12 hours.
  timeouts {
    create = "24h"
    update = "24h"
    delete = "12h"
  }
}
```

## Argument Reference

The following arguments are required:

* `az_mode` - (Required) Number of availability zones you want to assign per cluster. Valid values are `SINGLE` (assigns one availability zone per cluster) and `MULTI` (assigns all the availability zones per cluster).
* `database_name` - (Required) Name of the database where you want to create a dataview.
* `environment_id` - (Required) Unique identifier for the KX environment.
* `name` - (Required) Unique identifier for the dataview.

The following arguments are optional:

* `auto_update` - (Optional) Whether to apply all the future additions and corrections automatically to the dataview when you ingest new changesets. Defaults to `false`.
* `availability_zone_id` - (Optional) Identifier of the availability zones. If attaching a volume, the volume must be in the same availability zone as the dataview that you are attaching to.
* `changeset_id` - (Optional) Unique identifier of the changeset of the database that you want to use to ingest data.
* `description` - (Optional) Description for the dataview.
* `read_write` - (Optional) Whether to make the dataview writable to perform database maintenance. You cannot create partial writable dataviews; you must provide the entire database path and cannot perform updates, so `auto_update` must be `false` when `read_write` is `true`. You must also use a unique volume, and once a dataview is writable you cannot change it to read-only.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `segment_configurations` - (Optional) Configuration that contains the database path of the data that you want to place on each selected volume. Each segment must have a unique database path for each volume. If you do not explicitly specify any database path for a volume, they are accessible from the cluster through the default S3/object store segment. See [`segment_configurations`](#segment_configurations-block) below.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `segment_configurations` Block

* `db_paths` - (Required) Database path of the data that you want to place on each selected volume. Each segment must have a unique database path for each volume.
* `on_demand` - (Optional) Enables on-demand caching on the selected database path when a particular file or a column of the database is accessed. When on demand caching is **True**, dataviews perform minimal loading of files on the filesystem as needed. When it is set to **False**, everything is cached. The default value is **False**.
* `volume_name` - (Required) Name of the volume that you want to attach to a dataview. This volume must be in the same availability zone as the dataview that you are attaching to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX dataview.
* `created_timestamp` - Timestamp at which the dataview was created in FinSpace. Value determined as epoch time in milliseconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000000.
* `id` - Comma-delimited string joining environment ID, database name and dataview name.
* `last_modified_timestamp` - Last time that the dataview was updated in FinSpace. The value is determined as epoch time in milliseconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000000.
* `status` - Status of the dataview.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `4h`)
* `update` - (Default `4h`)
* `delete` - (Default `4h`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS FinSpace Kx Dataview using the `id` (environment ID, database name and dataview name, comma-delimited). For example:

```terraform
import {
  to = aws_finspace_kx_dataview.example
  id = "n3ceo7wqxoxcti5tujqwzs,my-tf-kx-database,my-tf-kx-dataview"
}
```

Using `terraform import`, import an AWS FinSpace Kx Cluster using the `id` (environment ID and cluster name, comma-delimited). For example:

```console
% terraform import aws_finspace_kx_dataview.example n3ceo7wqxoxcti5tujqwzs,my-tf-kx-database,my-tf-kx-dataview
```
