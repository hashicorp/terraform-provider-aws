---
layout: "aws"
page_title: "AWS: aws_elasticsearch_domain"
sidebar_current: "docs-aws-datasource-elasticsearch-domain"
description: |-
  Get information on an AWS Elasticsearch Domain resource.
---

# Data Source: aws_elasticsearch_domain

Use this data source to get information about an AWS Elasticsearch Domain.

## Example Usage

```hcl
data "aws_elasticsearch_domain" "bar" {
  domain_name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` – (Required) Name of the domain.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the domain.
* `domain_id` - Unique identifier for the domain.
* `domain_name` - The name of the Elasticsearch domain.
* `endpoint` - Domain-specific endpoint used to submit index, search, and data upload requests.
* `kibana_endpoint` - Domain-specific endpoint for kibana without https scheme.
* `vpc_options.0.availability_zones` - If the domain was created inside a VPC, the names of the availability zones the configured `subnet_ids` were created inside.
* `vpc_options.0.vpc_id` - If the domain was created inside a VPC, the ID of the VPC.
* `access_policies` - IAM policy document specifying the access policies for the domain
* `advanced_options` - Key-value string pairs to specify advanced configuration options.
   Note that the values for these configuration options must be strings (wrapped in quotes) or they
   may be wrong and cause a perpetual diff, causing Terraform to want to recreate your Elasticsearch
   domain on every apply.
* `ebs_options` - EBS related options, may be required based on chosen [instance size](https://aws.amazon.com/elasticsearch-service/pricing/). See below.
* `encrypt_at_rest` - Encrypt at rest options. Only available for [certain instance types](http://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/aes-supported-instance-types.html). See below.
* `node_to_node_encryption` - Node-to-node encryption options. See below.
* `cluster_config` - Cluster configuration of the domain, see below.
* `snapshot_options` - Snapshot related options, see below.
* `vpc_options` - VPC related options, see below. Adding or removing this configuration forces a new resource ([documentation](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-vpc.html#es-vpc-limitations)).
* `log_publishing_options` - Options for publishing slow logs to CloudWatch Logs.
* `elasticsearch_version` - The version of Elasticsearch to deploy. Defaults to `1.5`
* `tags` - A mapping of tags to assign to the resource

**ebs_options** the following attributes are exported:

* `ebs_enabled` - Whether EBS volumes are attached to data nodes in the domain
* `volume_type` - The type of EBS volumes attached to data nodes.
* `volume_size` - The size of EBS volumes attached to data nodes (in GB).
**Required** if `ebs_enabled` is set to `true`.
* `iops` - The baseline input/output (I/O) performance of EBS volumes
	attached to data nodes. Applicable only for the Provisioned IOPS EBS volume type.

**encrypt_at_rest** the following attributes are exported:

* `enabled` - Whether to enable encryption at rest. If the `encrypt_at_rest` block is not provided then this defaults to `false`.
* `kms_key_id` - The KMS key id to encrypt the Elasticsearch domain with. If not specified then it defaults to using the `aws/es` service KMS key.

**cluster_config** the following attributes are exported:

* `instance_type` - Instance type of data nodes in the cluster.
* `instance_count` - Number of instances in the cluster.
* `dedicated_master_enabled` - Indicates whether dedicated master nodes are enabled for the cluster.
* `dedicated_master_type` - Instance type of the dedicated master nodes in the cluster.
* `dedicated_master_count` - Number of dedicated master nodes in the cluster
* `zone_awareness_enabled` - Indicates whether zone awareness is enabled.

**node_to_node_encryption** the following attributes are exported:

* `enabled` - Whether to enable node-to-node encryption. If the `node_to_node_encryption` block is not provided then this defaults to `false`.

**vpc_options** the following attributes are exported:

AWS documentation: [VPC Support for Amazon Elasticsearch Service Domains](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-vpc.html)

**Note** you must have created the service linked role for the Elasticsearch service to use the `vpc_options`.
If you need to create the service linked role at the same time as the Elasticsearch domain then you must use `depends_on` to make sure that the role is created before the Elasticsearch domain.
See the [VPC based ES domain example](#vpc-based-es) above.

* `security_group_ids` - List of VPC Security Group IDs to be applied to the Elasticsearch domain endpoints. If omitted, the default Security Group for the VPC will be used.
* `subnet_ids` - List of VPC Subnet IDs for the Elasticsearch domain endpoints to be created in.

Security Groups and Subnets referenced in these attributes must all be within the same VPC; this determines what VPC the endpoints are created in.

**snapshot_options** the following attributes are exported:

* `automated_snapshot_start_hour` - Hour during which the service takes an automated daily
	snapshot of the indices in the domain.

**log_publishing_options** the following attributes are exported:

* `log_type` - A type of Elasticsearch log. Valid values: INDEX_SLOW_LOGS, SEARCH_SLOW_LOGS, ES_APPLICATION_LOGS
* `cloudwatch_log_group_arn` - ARN of the Cloudwatch log group to which log needs to be published.
* `enabled` - Specifies whether given log publishing option is enabled or not.

**cognito_options** the following attributes are exported:

AWS documentation: [Amazon Cognito Authentication for Kibana](https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-cognito-auth.html)

* `enabled` - Specifies whether Amazon Cognito authentication with Kibana is enabled or not
* `user_pool_id` - ID of the Cognito User Pool to use
* `identity_pool_id` - ID of the Cognito Identity Pool to use
* `role_arn` - ARN of the IAM role that has the AmazonESCognitoAccess policy attached