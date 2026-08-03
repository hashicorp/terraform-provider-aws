---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_cluster"
description: |-
  Terraform resource for managing an AWS FinSpace Kx Cluster.
---

# Resource: aws_finspace_kx_cluster

Terraform resource for managing an AWS FinSpace Kx Cluster.

## Example Usage

### Basic Usage

```terraform
resource "aws_finspace_kx_cluster" "example" {
  name                 = "my-tf-kx-cluster"
  environment_id       = aws_finspace_kx_environment.example.id
  type                 = "HDB"
  release_label        = "1.0"
  az_mode              = "SINGLE"
  availability_zone_id = "use1-az2"

  capacity_configuration {
    node_type  = "kx.s.2xlarge"
    node_count = 2
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = [aws_subnet.example.id]
    ip_address_type    = "IP_V4"
  }

  cache_storage_configurations {
    type = "CACHE_1000"
    size = 1200
  }

  database {
    database_name = aws_finspace_kx_database.example.name
    cache_configuration {
      cache_type = "CACHE_1000"
      db_paths   = "/"
    }
  }

  code {
    s3_bucket = aws_s3_bucket.test.id
    s3_key    = aws_s3_object.object.key
  }

  # Depending on the amount of data cached, create/update timeouts 
  # may need to be increased up to a potential maximum of 18 hours.
  timeouts {
    create = "18h"
    update = "18h"
  }
}
```

## Argument Reference

The following arguments are required:

* `az_mode` - (Required) Number of availability zones to assign per cluster. Valid values are `SINGLE` (assigns one availability zone per cluster) and `MULTI` (assigns all the availability zones per cluster).
* `capacity_configuration` - (Required) Structure for the metadata of a cluster. Includes information like the CPUs needed, memory of instances, and number of instances. See [`capacity_configuration` Block](#capacity_configuration-block).
* `environment_id` - (Required) Unique identifier for the KX environment.
* `name` - (Required) Unique name for the cluster that you want to create.
* `release_label` - (Required) Version of FinSpace Managed kdb to run.
* `type` - (Required) Type of KDB database. Valid values are `HDB` (Historical Database), `RDB` (Realtime Database, which requires the `savedown_storage_configuration` parameter), `GATEWAY`, `GP` (general purpose), and `Tickerplant`.
* `vpc_configuration` - (Required) Configuration details about the network where the Privatelink endpoint of the cluster resides. See [`vpc_configuration` Block](#vpc_configuration-block).

The following arguments are optional:

* `auto_scaling_configuration` - (Optional) Configuration based on which FinSpace will scale in or scale out nodes in your cluster. See [`auto_scaling_configuration` Block](#auto_scaling_configuration-block).
* `availability_zone_id` - (Optional) Availability zone identifiers for the requested regions. Required when `az_mode` is set to SINGLE.
* `cache_storage_configurations` - (Optional) Configurations for a read only cache storage associated with a cluster. This cache will be stored as an FSx Lustre that reads from the S3 store. See [`cache_storage_configurations` Block](#cache_storage_configurations-block).
* `code` - (Optional) Details of the custom code that you want to use inside a cluster when analyzing data. Consists of the S3 source bucket, location, object version, and the relative path from where the custom code is loaded into the cluster. See [`code` Block](#code-block).
* `command_line_arguments` - (Optional) List of key-value pairs to make available inside the cluster.
* `database` - (Optional) KX database that will be available for querying. See [`database` Block](#database-block).
* `description` - (Optional) Description of the cluster.
* `execution_role` - (Optional) IAM role that defines a set of permissions associated with a cluster. These permissions are assumed when a cluster attempts to access another cluster.
* `initialization_script` - (Optional) Path to Q program that will be run at launch of a cluster. This is a relative path within .zip file that contains the custom code, which will be loaded on the cluster. It must include the file name itself. For example, somedir/init.q.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `savedown_storage_configuration` - (Optional) Size and type of the temporary storage that is used to hold data during the savedown process. This parameter is required when you choose `type` as RDB. All the data written to this storage space is lost when the cluster node is restarted. See [`savedown_storage_configuration` Block](#savedown_storage_configuration-block).
* `scaling_group_configuration` - (Optional) Structure that stores the configuration details of a scaling group. See [`scaling_group_configuration` Block](#scaling_group_configuration-block).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tickerplant_log_configuration` - (Optional) Configuration to store Tickerplant logs. It consists of a list of volumes that will be mounted to your cluster. For the cluster type Tickerplant, the location of the TP volume on the cluster will be available by using the global variable .aws.tp_log_path. See [`tickerplant_log_configuration` Block](#tickerplant_log_configuration-block).

### `auto_scaling_configuration` Block

The `auto_scaling_configuration` block supports the following arguments:

* `auto_scaling_metric` - (Required) Metric your cluster will track in order to scale in and out. For example, CPU_UTILIZATION_PERCENTAGE is the average CPU usage across all nodes in a cluster.
* `max_node_count` - (Required) Highest number of nodes to scale. Cannot be greater than 5.
* `metric_target` - (Required) Desired value of chosen `auto_scaling_metric`. When metric drops below this value, cluster will scale in. When metric goes above this value, cluster will scale out. Can be set between 0 and 100 percent.
* `min_node_count` - (Required) Lowest number of nodes to scale. Must be at least 1 and less than the `max_node_count`. If nodes in cluster belong to multiple availability zones, then `min_node_count` must be at least 3.
* `scale_in_cooldown_seconds` - (Required) Duration in seconds that FinSpace will wait after a scale in event before initiating another scaling event.
* `scale_out_cooldown_seconds` - (Required) Duration in seconds that FinSpace will wait after a scale out event before initiating another scaling event.

### `capacity_configuration` Block

The `capacity_configuration` block supports the following arguments:

* `node_count` - (Required) Number of instances running in a cluster. Must be at least 1 and at most 5.
* `node_type` - (Required) Determines the hardware of the host computer used for your cluster instance. Valid values are `kx.s.large`, `kx.s.xlarge`, `kx.s.2xlarge`, `kx.s.4xlarge`, `kx.s.8xlarge`, `kx.s.16xlarge`, and `kx.s.32xlarge`.

### `cache_storage_configurations` Block

The `cache_storage_configurations` block supports the following arguments:

* `size` - (Required) Size of cache in Gigabytes.
* `type` - (Required) Type of cache storage. Valid values are `CACHE_1000` (1000 MB/s disk access throughput), `CACHE_250` (250 MB/s disk access throughput), and `CACHE_12` (12 MB/s disk access throughput).

### `code` Block

The `code` block supports the following arguments:

* `s3_bucket` - (Required) Unique name for the S3 bucket.
* `s3_key` - (Required) Full S3 path (excluding bucket) to the .zip file that contains the code to be loaded onto the cluster when it’s started.
* `s3_object_version` - (Optional) Version of an S3 Object.

### `database` Block

The `database` block supports the following arguments:

* `cache_configurations` - (Optional) Configuration details for the disk cache to increase performance reading from a KX database mounted to the cluster. See [`cache_configurations` Block](#cache_configurations-block).
* `changeset_id` - (Optional) Unique identifier of the changeset that is associated with the cluster.
* `database_name` - (Required) Name of the KX database.
* `dataview_name` - (Optional) Name of the dataview to be used for caching historical data on disk. You cannot update to a different dataview name once a cluster is created. Use `lifecycle` [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) for database to prevent any undesirable behaviors.

#### `cache_configurations` Block

The `cache_configurations` block supports the following arguments:

* `cache_type` - (Required) Type of disk cache.
* `db_paths` - (Optional) Paths within the database to cache.

### `savedown_storage_configuration` Block

The `savedown_storage_configuration` block supports the following arguments:

* `size` - (Optional) Size of temporary storage in gigabytes. Must be between 10 and 16000.
* `type` - (Optional) Type of writeable storage space for temporarily storing your savedown data. Valid value is `SDS01`, which represents 3000 IOPS and io2 ebs volume type.
* `volume_name` - (Optional) Name of the kdb volume that you want to use as writeable save-down storage for clusters.

### `vpc_configuration` Block

The `vpc_configuration` block supports the following arguments:

* `ip_address_type` - (Required) IP address type for cluster network configuration parameters. The following type is available: IP_V4 - IP address version 4.
* `security_group_ids` - (Required) Unique identifier of the VPC security group applied to the VPC endpoint ENI for the cluster.
* `subnet_ids` - (Required) Identifier of the subnet that the Privatelink VPC endpoint uses to connect to the cluster.
* `vpc_id` - (Required) Identifier of the VPC endpoint.

### `scaling_group_configuration` Block

The `scaling_group_configuration` block supports the following arguments:

* `cpu` - (Optional) Number of vCPUs that you want to reserve for each node of this kdb cluster on the scaling group host.
* `memory_limit` - (Optional) Hard limit on the amount of memory a kdb cluster can use.
* `memory_reservation` - (Required) Reservation of the minimum amount of memory that should be available on the scaling group for a kdb cluster to be successfully placed in a scaling group.
* `node_count` - (Required) Number of kdb cluster nodes.
* `scaling_group_name` - (Required) Unique identifier for the kdb scaling group.

### `tickerplant_log_configuration` Block

The `tickerplant_log_configuration` block supports the following arguments:

* `tickerplant_log_volumes` - (Required) Names of the volumes for tickerplant logs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX cluster.
* `created_timestamp` - Timestamp at which the cluster is created in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `id` - Comma-delimited string joining environment ID and cluster name.
* `last_modified_timestamp` - Last timestamp at which the cluster was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `status` - Status of the cluster.
* `status_reason` - Reason for the cluster status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `4h`)
* `update` - (Default `4h`)
* `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS FinSpace Kx Cluster using the `id` (environment ID and cluster name, comma-delimited). For example:

```terraform
import {
  to = aws_finspace_kx_cluster.example
  id = "n3ceo7wqxoxcti5tujqwzs,my-tf-kx-cluster"
}
```

Using `terraform import`, import an AWS FinSpace Kx Cluster using the `id` (environment ID and cluster name, comma-delimited). For example:

```console
% terraform import aws_finspace_kx_cluster.example n3ceo7wqxoxcti5tujqwzs,my-tf-kx-cluster
```
