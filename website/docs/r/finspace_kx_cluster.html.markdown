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

* `az_mode` - (Required) The number of availability zones you want to assign per cluster. This can be one of the following:
    * SINGLE - Assigns one availability zone per cluster.
    * MULTI - Assigns all the availability zones per cluster.
* `capacity_configuration` - (Required) Structure for the metadata of a cluster. Includes information like the CPUs needed, memory of instances, and number of instances. See [capacity_configuration](#capacity_configuration).
* `environment_id` - (Required) Unique identifier for the KX environment.
* `name` - (Required) Unique name for the cluster that you want to create.
* `release_label` - (Required) Version of FinSpace Managed kdb to run.
* `type` - (Required) Type of KDB database. The following types are available:
    * HDB - Historical Database. The data is only accessible with read-only permissions from one of the FinSpace managed KX databases mounted to the cluster.
    * RDB - Realtime Database. This type of database captures all the data from a ticker plant and stores it in memory until the end of day, after which it writes all of its data to a disk and reloads the HDB. This cluster type requires local storage for temporary storage of data during the savedown process. If you specify this field in your request, you must provide the `savedownStorageConfiguration` parameter.
    * GATEWAY - A gateway cluster allows you to access data across processes in kdb systems. It allows you to create your own routing logic using the initialization scripts and custom code. This type of cluster does not require a  writable local storage.
    * GP - A general purpose cluster allows you to quickly iterate on code during development by granting greater access to system commands and enabling a fast reload of custom code. This cluster type can optionally mount databases including cache and savedown storage. For this cluster type, the node count is fixed at 1. It does not support autoscaling and supports only `SINGLE` AZ mode.
    * Tickerplant – A tickerplant cluster allows you to subscribe to feed handlers based on IAM permissions. It can publish to RDBs, other Tickerplants, and real-time subscribers (RTS). Tickerplants can persist messages to log, which is readable by any RDB environment. It supports only single-node that is only one kdb process.
* `vpc_configuration` - (Required) Configuration details about the network where the Privatelink endpoint of the cluster resides. See [vpc_configuration](#vpc_configuration).

The following arguments are optional:

* `auto_scaling_configuration` - (Optional) Configuration based on which FinSpace will scale in or scale out nodes in your cluster. See [auto_scaling_configuration](#auto_scaling_configuration).
* `availability_zone_id` - (Optional) The availability zone identifiers for the requested regions. Required when `az_mode` is set to SINGLE.
* `cache_storage_configurations` - (Optional) Configurations for a read only cache storage associated with a cluster. This cache will be stored as an FSx Lustre that reads from the S3 store. See [cache_storage_configuration](#cache_storage_configuration).
* `code` - (Optional) Details of the custom code that you want to use inside a cluster when analyzing data. Consists of the S3 source bucket, location, object version, and the relative path from where the custom code is loaded into the cluster. See [code](#code).
* `command_line_arguments` - (Optional) List of key-value pairs to make available inside the cluster.
* `database` - (Optional) KX database that will be available for querying. Defined below.
* `description` - (Optional) Description of the cluster.
* `execution_role` - (Optional) An IAM role that defines a set of permissions associated with a cluster. These permissions are assumed when a cluster attempts to access another cluster.
* `initialization_script` - (Optional) Path to Q program that will be run at launch of a cluster. This is a relative path within .zip file that contains the custom code, which will be loaded on the cluster. It must include the file name itself. For example, somedir/init.q.
* `savedown_storage_configuration` - (Optional) Size and type of the temporary storage that is used to hold data during the savedown process. This parameter is required when you choose `type` as RDB. All the data written to this storage space is lost when the cluster node is restarted. See [savedown_storage_configuration](#savedown_storage_configuration).
* `scaling_group_configuration` - (Optional) The structure that stores the configuration details of a scaling group.
* `tickerplant_log_configuration` - A configuration to store Tickerplant logs. It consists of a list of volumes that will be mounted to your cluster. For the cluster type Tickerplant , the location of the TP volume on the cluster will be available by using the global variable .aws.tp_log_path.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### auto_scaling_configuration

The auto_scaling_configuration block supports the following arguments:

* `auto_scaling_metric` - (Required) Metric your cluster will track in order to scale in and out. For example, CPU_UTILIZATION_PERCENTAGE is the average CPU usage across all nodes in a cluster.
* `min_node_count` - (Required) Lowest number of nodes to scale. Must be at least 1 and less than the `max_node_count`. If nodes in cluster belong to multiple availability zones, then `min_node_count` must be at least 3.
* `max_node_count` - (Required) Highest number of nodes to scale. Cannot be greater than 5
* `metric_target` - (Required) Desired value of chosen `auto_scaling_metric`. When metric drops below this value, cluster will scale in. When metric goes above this value, cluster will scale out. Can be set between 0 and 100 percent.
* `scale_in_cooldown_seconds` - (Required) Duration in seconds that FinSpace will wait after a scale in event before initiating another scaling event.
* `scale_out_cooldown_seconds` - (Required) Duration in seconds that FinSpace will wait after a scale out event before initiating another scaling event.

### capacity_configuration

The capacity_configuration block supports the following arguments:

* `node_type` - (Required) Determines the hardware of the host computer used for your cluster instance. Each node type offers different memory and storage capabilities. Choose a node type based on the requirements of the application or software that you plan to run on your instance.
  
  You can only specify one of the following values:
    * kx.s.large – The node type with a configuration of 12 GiB memory and 2 vCPUs.
    * kx.s.xlarge – The node type with a configuration of 27 GiB memory and 4 vCPUs.
    * kx.s.2xlarge – The node type with a configuration of 54 GiB memory and 8 vCPUs.
    * kx.s.4xlarge – The node type with a configuration of 108 GiB memory and 16 vCPUs.
    * kx.s.8xlarge – The node type with a configuration of 216 GiB memory and 32 vCPUs.
    * kx.s.16xlarge – The node type with a configuration of 432 GiB memory and 64 vCPUs.
    * kx.s.32xlarge – The node type with a configuration of 864 GiB memory and 128 vCPUs.
* `node_count` - (Required) Number of instances running in a cluster. Must be at least 1 and at most 5.

### cache_storage_configuration

The cache_storage_configuration block supports the following arguments:

* `type` - (Required) Type of cache storage . The valid values are:
    * CACHE_1000 - This type provides 1000 MB/s disk access throughput.
    * CACHE_250 - This type provides 250 MB/s disk access throughput.
    * CACHE_12 - This type provides 12 MB/s disk access throughput.
* `size` - (Required) Size of cache in Gigabytes.

Please note that create/update timeouts may have to be adjusted from the default 4 hours depending upon the
volume of data being cached, as noted in the example configuration.

### code

The code block supports the following arguments:

* `s3_bucket` - (Required) Unique name for the S3 bucket.
* `s3_key` - (Required) Full S3 path (excluding bucket) to the .zip file that contains the code to be loaded onto the cluster when it’s started.
* `s3_object_version` - (Optional) Version of an S3 Object.

### database

The database block supports the following arguments:

* `database_name` - (Required) Name of the KX database.
* `cache_configurations` - (Optional) Configuration details for the disk cache to increase performance reading from a KX database mounted to the cluster. See [cache_configurations](#cache_configurations).
* `changeset_id` - (Optional) A unique identifier of the changeset that is associated with the cluster.
* `dataview_name` - (Optional) The name of the dataview to be used for caching historical data on disk. You cannot update to a different dataview name once a cluster is created. Use `lifecycle` [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) for database to prevent any undesirable behaviors.

#### cache_configurations

The cache_configuration block supports the following arguments:

* `cache_type` - (Required) Type of disk cache.
* `db_paths` - (Optional) Paths within the database to cache.

### savedown_storage_configuration

The savedown_storage_configuration block supports the following arguments:

* `type` - (Optional) Type of writeable storage space for temporarily storing your savedown data. The valid values are:
    * SDS01 - This type represents 3000 IOPS and io2 ebs volume type.
* `size` - (Optional) Size of temporary storage in gigabytes. Must be between 10 and 16000.
* `volume_name` - (Optional) The name of the kdb volume that you want to use as writeable save-down storage for clusters.

### vpc_configuration

The vpc_configuration block supports the following arguments:

* `vpc_id` - (Required) Identifier of the VPC endpoint
* `security_group_ids` - (Required) Unique identifier of the VPC security group applied to the VPC endpoint ENI for the cluster.
* `subnet_ids `- (Required) Identifier of the subnet that the Privatelink VPC endpoint uses to connect to the cluster.
* `ip_address_type` - (Required) IP address type for cluster network configuration parameters. The following type is available: IP_V4 - IP address version 4.

### scaling_group_configuration

The scaling_group_configuration block supports the following arguments:

* `scaling_group_name` - (Required) A unique identifier for the kdb scaling group.
* `memory_reservation` - (Required) A reservation of the minimum amount of memory that should be available on the scaling group for a kdb cluster to be successfully placed in a scaling group.
* `node_count` - (Required) The number of kdb cluster nodes.
* `cpu` - The number of vCPUs that you want to reserve for each node of this kdb cluster on the scaling group host.
* `memory_limit` - An optional hard limit on the amount of memory a kdb cluster can use.

### tickerplant_log_configuration

The tickerplant_log_configuration block supports the following arguments:

* tickerplant_log_volumes - (Required) The names of the volumes for tickerplant logs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX cluster.
* `created_timestamp` - Timestamp at which the cluster is created in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `id` - A comma-delimited string joining environment ID and cluster name.
* `last_modified_timestamp` - Last timestamp at which the cluster was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
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
