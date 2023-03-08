---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_domain"
description: |-
  Get information on an OpenSearch Domain resource.
---

# Data Source: aws_opensearch_domain

Use this data source to get information about an OpenSearch Domain

## Example Usage

```terraform
data "aws_opensearch_domain" "my_domain" {
  domain_name = "my-domain-name"
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` – (Required) Name of the domain.

## Attributes Reference

The following attributes are exported:

* `access_policies` – Policy document attached to the domain.
* `advanced_options` - Key-value string pairs to specify advanced configuration options.
* `advanced_security_options` - Status of the OpenSearch domain's advanced security options. The block consists of the following attributes:
    * `enabled` - Whether advanced security is enabled.
    * `internal_user_database_enabled` - Whether the internal user database is enabled.
* `arn` – ARN of the domain.
* `auto_tune_options` - Configuration of the Auto-Tune options of the domain.
    * `desired_state` - Auto-Tune desired state for the domain.
    * `maintenance_schedule` - A list of the nested configurations for the Auto-Tune maintenance windows of the domain.
        * `start_at` - Date and time at which the Auto-Tune maintenance schedule starts in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
        * `duration` - Configuration block for the duration of the Auto-Tune maintenance window.
            * `value` - Duration of an Auto-Tune maintenance window.
            * `unit` - Unit of time.
        * `cron_expression_for_recurrence` - Cron expression for an Auto-Tune maintenance schedule.
    * `rollback_on_disable` - Whether the domain is set to roll back to default Auto-Tune settings when disabling Auto-Tune.
* `cluster_config` - Cluster configuration of the domain.
    * `cold_storage_options` - Configuration block containing cold storage configuration.
        * `enabled` - Indicates  cold storage is enabled.
    * `instance_type` - Instance type of data nodes in the cluster.
    * `instance_count` - Number of instances in the cluster.
    * `dedicated_master_enabled` - Indicates whether dedicated master nodes are enabled for the cluster.
    * `dedicated_master_type` - Instance type of the dedicated master nodes in the cluster.
    * `dedicated_master_count` - Number of dedicated master nodes in the cluster.
    * `zone_awareness_enabled` - Indicates whether zone awareness is enabled.
    * `zone_awareness_config` - Configuration block containing zone awareness settings.
        * `availability_zone_count` - Number of availability zones used.
    * `warm_enabled` - Warm storage is enabled.
    * `warm_count` - Number of warm nodes in the cluster.
    * `warm_type` - Instance type for the OpenSearch cluster's warm nodes.
* `cognito_options` - Domain Amazon Cognito Authentication options for Kibana.
    * `enabled` - Whether Amazon Cognito Authentication is enabled.
    * `user_pool_id` - Cognito User pool used by the domain.
    * `identity_pool_id` - Cognito Identity pool used by the domain.
    * `role_arn` - IAM Role with the AmazonOpenSearchServiceCognitoAccess policy attached.
* `created` – Status of the creation of the domain.
* `deleted` – Status of the deletion of the domain.
* `domain_id` – Unique identifier for the domain.
* `ebs_options` - EBS Options for the instances in the domain.
    * `ebs_enabled` - Whether EBS volumes are attached to data nodes in the domain.
    * `throughput` - The throughput (in MiB/s) of the EBS volumes attached to data nodes.
    * `volume_type` - Type of EBS volumes attached to data nodes.
    * `volume_size` - Size of EBS volumes attached to data nodes (in GB).
    * `iops` - Baseline input/output (I/O) performance of EBS volumes attached to data nodes.
* `engine_version` – OpenSearch version for the domain.
* `encryption_at_rest` - Domain encryption at rest related options.
    * `enabled` - Whether encryption at rest is enabled in the domain.
    * `kms_key_id` - KMS key id used to encrypt data at rest.
* `endpoint` – Domain-specific endpoint used to submit index, search, and data upload requests.
* `kibana_endpoint` - Domain-specific endpoint used to access the Kibana application.
* `log_publishing_options` - Domain log publishing related options.
    * `log_type` - Type of OpenSearch log being published.
    * `cloudwatch_log_group_arn` - CloudWatch Log Group where the logs are published.
    * `enabled` - Whether log publishing is enabled.
* `node_to_node_encryption` - Domain in transit encryption related options.
    * `enabled` - Whether node to node encryption is enabled.
* `processing` – Status of a configuration change in the domain.
* `snapshot_options` – Domain snapshot related options.
    * `automated_snapshot_start_hour` - Hour during which the service takes an automated daily snapshot of the indices in the domain.
* `tags` - Tags assigned to the domain.
* `vpc_options` - VPC Options for private OpenSearch domains.
    * `availability_zones` - Availability zones used by the domain.
    * `security_group_ids` - Security groups used by the domain.
    * `subnet_ids` - Subnets used by the domain.
    * `vpc_id` - VPC used by the domain.
