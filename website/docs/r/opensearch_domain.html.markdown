---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_domain"
description: |-
  Terraform resource for managing an AWS OpenSearch Domain.
---

# Resource: aws_opensearch_domain

Manages an Amazon OpenSearch Domain.

## Elasticsearch vs. OpenSearch

Amazon OpenSearch Service is the successor to Amazon Elasticsearch Service and supports OpenSearch and legacy Elasticsearch OSS (up to 7.10, the final open source version of the software).

OpenSearch Domain configurations are similar in many ways to Elasticsearch Domain configurations. However, there are important differences including these:

* OpenSearch has `engine_version` while Elasticsearch has `elasticsearch_version`
* Versions are specified differently - _e.g._, `Elasticsearch_7.10` with OpenSearch vs. `7.10` for Elasticsearch.
* `instance_type` argument values end in `search` for OpenSearch vs. `elasticsearch` for Elasticsearch (_e.g._, `t2.micro.search` vs. `t2.micro.elasticsearch`).
* The AWS-managed service-linked role for OpenSearch is called `AWSServiceRoleForAmazonOpenSearchService` instead of `AWSServiceRoleForAmazonElasticsearchService` for Elasticsearch.

There are also some potentially unexpected similarities in configurations:

* ARNs for both are prefaced with `arn:aws:es:`.
* Both OpenSearch and Elasticsearch use assume role policies that refer to the `Principal` `Service` as `es.amazonaws.com`.
* IAM policy actions, such as those you will find in `access_policies`, are prefaced with `es:` for both.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearch_domain" "example" {
  domain_name    = "example"
  engine_version = "Elasticsearch_7.10"

  cluster_config {
    instance_type = "r4.large.search"
  }

  tags = {
    Domain = "TestDomain"
  }
}
```

### Access Policy

-> See also: [`aws_opensearch_domain_policy` resource](/docs/providers/aws/r/opensearch_domain_policy.html)

```terraform
variable "domain" {
  default = "tf-test"
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["es:*"]
    resources = ["arn:aws:es:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:domain/${var.domain}/*"]

    condition {
      test     = "IpAddress"
      variable = "aws:SourceIp"
      values   = ["66.193.100.22/32"]
    }
  }
}

resource "aws_opensearch_domain" "example" {
  domain_name = var.domain

  # ... other configuration ...

  access_policies = data.aws_iam_policy_document.example.json
}
```

### Log publishing to CloudWatch Logs

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["es.amazonaws.com"]
    }

    actions = [
      "logs:PutLogEvents",
      "logs:PutLogEventsBatch",
      "logs:CreateLogStream",
    ]

    resources = ["arn:aws:logs:*"]
  }
}
resource "aws_cloudwatch_log_resource_policy" "example" {
  policy_name     = "example"
  policy_document = data.aws_iam_policy_document.example.json
}

resource "aws_opensearch_domain" "example" {
  # .. other configuration ...

  log_publishing_options {
    cloudwatch_log_group_arn = aws_cloudwatch_log_group.example.arn
    log_type                 = "INDEX_SLOW_LOGS"
  }
}
```

### VPC based OpenSearch

```terraform
variable "vpc" {}

variable "domain" {
  default = "tf-test"
}

data "aws_vpc" "example" {
  tags = {
    Name = var.vpc
  }
}

data "aws_subnets" "example" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.example.id]
  }

  tags = {
    Tier = "private"
  }
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_security_group" "example" {
  name        = "${var.vpc}-opensearch-${var.domain}"
  description = "Managed by Terraform"
  vpc_id      = data.aws_vpc.example.id

  ingress {
    from_port = 443
    to_port   = 443
    protocol  = "tcp"

    cidr_blocks = [
      data.aws_vpc.example.cidr_block,
    ]
  }
}

resource "aws_iam_service_linked_role" "example" {
  aws_service_name = "opensearchservice.amazonaws.com"
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["es:*"]
    resources = ["arn:aws:es:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:domain/${var.domain}/*"]
  }
}

resource "aws_opensearch_domain" "example" {
  domain_name    = var.domain
  engine_version = "OpenSearch_1.0"

  cluster_config {
    instance_type          = "m4.large.search"
    zone_awareness_enabled = true
  }

  vpc_options {
    subnet_ids = [
      data.aws_subnets.example.ids[0],
      data.aws_subnets.example.ids[1],
    ]

    security_group_ids = [aws_security_group.example.id]
  }

  advanced_options = {
    "rest.action.multi.allow_explicit_index" = "true"
  }

  access_policies = data.aws_iam_policy_document.example.json

  tags = {
    Domain = "TestDomain"
  }

  depends_on = [aws_iam_service_linked_role.example]
}
```

### Enabling fine-grained access control on an existing domain

This example shows two configurations: one to create a domain without fine-grained access control and the second to modify the domain to enable fine-grained access control. For more information, see [Enabling fine-grained access control](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/fgac.html).

#### First apply

```terraform
resource "aws_opensearch_domain" "example" {
  domain_name    = "ggkitty"
  engine_version = "Elasticsearch_7.1"

  cluster_config {
    instance_type = "r5.large.search"
  }

  advanced_security_options {
    enabled                        = false
    anonymous_auth_enabled         = true
    internal_user_database_enabled = true
    master_user_options {
      master_user_name     = "example"
      master_user_password = "Barbarbarbar1!"
    }
  }

  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
```

#### Second apply

Notice that the only change is `advanced_security_options.0.enabled` is now set to `true`.

```terraform
resource "aws_opensearch_domain" "example" {
  domain_name    = "ggkitty"
  engine_version = "Elasticsearch_7.1"

  cluster_config {
    instance_type = "r5.large.search"
  }

  advanced_security_options {
    enabled                        = true
    anonymous_auth_enabled         = true
    internal_user_database_enabled = true
    master_user_options {
      master_user_name     = "example"
      master_user_password = "Barbarbarbar1!"
    }
  }

  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_name` - (Required) Name of the domain.

The following arguments are optional:

* `access_policies` - (Optional) IAM policy document specifying the access policies for the domain.
* `advanced_options` - (Optional) Key-value string pairs to specify advanced configuration options. Note that the values for these configuration options must be strings (wrapped in quotes) or they may be wrong and cause a perpetual diff, causing Terraform to want to recreate your OpenSearch domain on every apply.
* `advanced_security_options` - (Optional) Configuration block for [fine-grained access control](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/fgac.html). Detailed below.
* `auto_tune_options` - (Optional) Configuration block for the Auto-Tune options of the domain. Detailed below.
* `cluster_config` - (Optional) Configuration block for the cluster of the domain. Detailed below.
* `cognito_options` - (Optional) Configuration block for authenticating dashboard with Cognito. Detailed below.
* `domain_endpoint_options` - (Optional) Configuration block for domain endpoint HTTP(S) related options. Detailed below.
* `ebs_options` - (Optional) Configuration block for EBS related options, may be required based on chosen [instance size](https://aws.amazon.com/opensearch-service/pricing/). Detailed below.
* `engine_version` - (Optional) Either `Elasticsearch_X.Y` or `OpenSearch_X.Y` to specify the engine version for the Amazon OpenSearch Service domain. For example, `OpenSearch_1.0` or `Elasticsearch_7.9`.
  See [Creating and managing Amazon OpenSearch Service domains](http://docs.aws.amazon.com/opensearch-service/latest/developerguide/createupdatedomains.html#createdomains).
  Defaults to the lastest version of OpenSearch.
* `ip_address_type` - (Optional) The IP address type for the endpoint. Valid values are `ipv4` and `dualstack`.
* `encrypt_at_rest` - (Optional) Configuration block for encrypt at rest options. Only available for [certain instance types](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/encryption-at-rest.html). Detailed below.
* `log_publishing_options` - (Optional) Configuration block for publishing slow and application logs to CloudWatch Logs. This block can be declared multiple times, for each log_type, within the same resource. Detailed below.
* `node_to_node_encryption` - (Optional) Configuration block for node-to-node encryption options. Detailed below.
* `snapshot_options` - (Optional) Configuration block for snapshot related options. Detailed below. DEPRECATED. For domains running OpenSearch 5.3 and later, Amazon OpenSearch takes hourly automated snapshots, making this setting irrelevant. For domains running earlier versions, OpenSearch takes daily automated snapshots.
* `software_update_options` - (Optional) Software update options for the domain. Detailed below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_options` - (Optional) Configuration block for VPC related options. Adding or removing this configuration forces a new resource ([documentation](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/vpc.html)). Detailed below.
* `off_peak_window_options` - (Optional) Configuration to add Off Peak update options. ([documentation](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/off-peak.html)). Detailed below.

### advanced_security_options

* `anonymous_auth_enabled` - (Optional) Whether Anonymous auth is enabled. Enables fine-grained access control on an existing domain. Ignored unless `advanced_security_options` are enabled. _Can only be enabled on an existing domain._
* `enabled` - (Required, Forces new resource when changing from `true` to `false`) Whether advanced security is enabled.
* `internal_user_database_enabled` - (Optional) Whether the internal user database is enabled. Default is `false`.
* `master_user_options` - (Optional) Configuration block for the main user. Detailed below.

#### master_user_options

* `master_user_arn` - (Optional) ARN for the main user. Only specify if `internal_user_database_enabled` is not set or set to `false`.
* `master_user_name` - (Optional) Main user's username, which is stored in the Amazon OpenSearch Service domain's internal database. Only specify if `internal_user_database_enabled` is set to `true`.
* `master_user_password` - (Optional) Main user's password, which is stored in the Amazon OpenSearch Service domain's internal database. Only specify if `internal_user_database_enabled` is set to `true`.

### auto_tune_options

* `desired_state` - (Required) Auto-Tune desired state for the domain. Valid values: `ENABLED` or `DISABLED`.
* `maintenance_schedule` - (Required if `rollback_on_disable` is set to `DEFAULT_ROLLBACK`) Configuration block for Auto-Tune maintenance windows. Can be specified multiple times for each maintenance window. Detailed below.

  **NOTE:** Maintenance windows are deprecated and have been replaced with [off-peak windows](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/off-peak.html). Consequently, `maintenance_schedule` configuration blocks cannot be specified when `use_off_peak_window` is set to `true`.
* `rollback_on_disable` - (Optional) Whether to roll back to default Auto-Tune settings when disabling Auto-Tune. Valid values: `DEFAULT_ROLLBACK` or `NO_ROLLBACK`.
* `use_off_peak_window` - (Optional) Whether to schedule Auto-Tune optimizations that require blue/green deployments during the domain's configured daily off-peak window. Defaults to `false`.

#### maintenance_schedule

* `start_at` - (Required) Date and time at which to start the Auto-Tune maintenance schedule in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `duration` - (Required) Configuration block for the duration of the Auto-Tune maintenance window. Detailed below.
* `cron_expression_for_recurrence` - (Required) A cron expression specifying the recurrence pattern for an Auto-Tune maintenance schedule.

##### duration

* `value` - (Required) An integer specifying the value of the duration of an Auto-Tune maintenance window.
* `unit` - (Required) Unit of time specifying the duration of an Auto-Tune maintenance window. Valid values: `HOURS`.

### cluster_config

* `cold_storage_options` - (Optional) Configuration block containing cold storage configuration. Detailed below.
* `dedicated_master_count` - (Optional) Number of dedicated main nodes in the cluster.
* `dedicated_master_enabled` - (Optional) Whether dedicated main nodes are enabled for the cluster.
* `dedicated_master_type` - (Optional) Instance type of the dedicated main nodes in the cluster.
* `instance_count` - (Optional) Number of instances in the cluster.
* `instance_type` - (Optional) Instance type of data nodes in the cluster.
* `multi_az_with_standby_enabled` - (Optional) Whether a multi-AZ domain is turned on with a standby AZ. For more information, see [Configuring a multi-AZ domain in Amazon OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/managedomains-multiaz.html).
* `warm_count` - (Optional) Number of warm nodes in the cluster. Valid values are between `2` and `150`. `warm_count` can be only and must be set when `warm_enabled` is set to `true`.
* `warm_enabled` - (Optional) Whether to enable warm storage.
* `warm_type` - (Optional) Instance type for the OpenSearch cluster's warm nodes. Valid values are `ultrawarm1.medium.search`, `ultrawarm1.large.search` and `ultrawarm1.xlarge.search`. `warm_type` can be only and must be set when `warm_enabled` is set to `true`.
* `zone_awareness_config` - (Optional) Configuration block containing zone awareness settings. Detailed below.
* `zone_awareness_enabled` - (Optional) Whether zone awareness is enabled, set to `true` for multi-az deployment. To enable awareness with three Availability Zones, the `availability_zone_count` within the `zone_awareness_config` must be set to `3`.

#### cold_storage_options

* `enabled` - (Optional) Boolean to enable cold storage for an OpenSearch domain. Defaults to `false`. Master and ultrawarm nodes must be enabled for cold storage.

#### zone_awareness_config

* `availability_zone_count` - (Optional) Number of Availability Zones for the domain to use with `zone_awareness_enabled`. Defaults to `2`. Valid values: `2` or `3`.

### cognito_options

AWS documentation: [Amazon Cognito Authentication for Dashboard](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/es-cognito-auth.html)

* `enabled` - (Optional) Whether Amazon Cognito authentication with Dashboard is enabled or not. Default is `false`.
* `identity_pool_id` - (Required) ID of the Cognito Identity Pool to use.
* `role_arn` - (Required) ARN of the IAM role that has the AmazonOpenSearchServiceCognitoAccess policy attached.
* `user_pool_id` - (Required) ID of the Cognito User Pool to use.

### domain_endpoint_options

* `custom_endpoint_certificate_arn` - (Optional) ACM certificate ARN for your custom endpoint.
* `custom_endpoint_enabled` - (Optional) Whether to enable custom endpoint for the OpenSearch domain.
* `custom_endpoint` - (Optional) Fully qualified domain for your custom endpoint.
* `enforce_https` - (Optional) Whether or not to require HTTPS. Defaults to `true`.
* `tls_security_policy` - (Optional) Name of the TLS security policy that needs to be applied to the HTTPS endpoint. For valid values, refer to the [AWS documentation](https://docs.aws.amazon.com/opensearch-service/latest/APIReference/API_DomainEndpointOptions.html#opensearchservice-Type-DomainEndpointOptions-TLSSecurityPolicy). Terraform will only perform drift detection if a configuration value is provided.

### ebs_options

* `ebs_enabled` - (Required) Whether EBS volumes are attached to data nodes in the domain.
* `iops` - (Optional) Baseline input/output (I/O) performance of EBS volumes attached to data nodes. Applicable only for the GP3 and Provisioned IOPS EBS volume types.
* `throughput` - (Required if `volume_type` is set to `gp3`) Specifies the throughput (in MiB/s) of the EBS volumes attached to data nodes. Applicable only for the gp3 volume type.
* `volume_size` - (Required if `ebs_enabled` is set to `true`.) Size of EBS volumes attached to data nodes (in GiB).
* `volume_type` - (Optional) Type of EBS volumes attached to data nodes.

### encrypt_at_rest

~> **Note:** You can enable `encrypt_at_rest` _in place_ for an existing, unencrypted domain only if you are using OpenSearch or your Elasticsearch version is 6.7 or greater. For other versions, if you enable `encrypt_at_rest`, Terraform with recreate the domain, potentially causing data loss. For any version, if you disable `encrypt_at_rest` for an existing, encrypted domain, Terraform will recreate the domain, potentially causing data loss. If you change the `kms_key_id`, Terraform will also recreate the domain, potentially causing data loss.

* `enabled` - (Required) Whether to enable encryption at rest. If the `encrypt_at_rest` block is not provided then this defaults to `false`. Enabling encryption on new domains requires an `engine_version` of `OpenSearch_X.Y` or `Elasticsearch_5.1` or greater.
* `kms_key_id` - (Optional) KMS key ARN to encrypt the Elasticsearch domain with. If not specified then it defaults to using the `aws/es` service KMS key. Note that KMS will accept a KMS key ID but will return the key ARN. To prevent Terraform detecting unwanted changes, use the key ARN instead.

### log_publishing_options

* `cloudwatch_log_group_arn` - (Required) ARN of the Cloudwatch log group to which log needs to be published.
* `enabled` - (Optional, Default: true) Whether given log publishing option is enabled or not.
* `log_type` - (Required) Type of OpenSearch log. Valid values: `INDEX_SLOW_LOGS`, `SEARCH_SLOW_LOGS`, `ES_APPLICATION_LOGS`, `AUDIT_LOGS`.

### node_to_node_encryption

~> **Note:** You can enable `node_to_node_encryption` _in place_ for an existing, unencrypted domain only if you are using OpenSearch or your Elasticsearch version is 6.7 or greater. For other versions, if you enable `node_to_node_encryption`, Terraform will recreate the domain, potentially causing data loss. For any version, if you disable `node_to_node_encryption` for an existing, node-to-node encrypted domain, Terraform will recreate the domain, potentially causing data loss.

* `enabled` - (Required) Whether to enable node-to-node encryption. If the `node_to_node_encryption` block is not provided then this defaults to `false`. Enabling node-to-node encryption of a new domain requires an `engine_version` of `OpenSearch_X.Y` or `Elasticsearch_6.0` or greater.

### snapshot_options

* `automated_snapshot_start_hour` - (Required) Hour during which the service takes an automated daily snapshot of the indices in the domain.

### software_update_options

* `auto_software_update_enabled` - (Optional) Whether automatic service software updates are enabled for the domain. Defaults to `false`.

### vpc_options

AWS documentation: [VPC Support for Amazon OpenSearch Service Domains](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/es-vpc.html)

~> **Note:** You must have created the service linked role for the OpenSearch service to use `vpc_options`. If you need to create the service linked role at the same time as the OpenSearch domain then you must use `depends_on` to make sure that the role is created before the OpenSearch domain. See the [VPC based ES domain example](#vpc-based-opensearch) above.

-> Security Groups and Subnets referenced in these attributes must all be within the same VPC. This determines what VPC the endpoints are created in.

* `security_group_ids` - (Optional) List of VPC Security Group IDs to be applied to the OpenSearch domain endpoints. If omitted, the default Security Group for the VPC will be used.
* `subnet_ids` - (Required) List of VPC Subnet IDs for the OpenSearch domain endpoints to be created in.

### off_peak_window_options

AWS documentation: [Off Peak Hours Support for Amazon OpenSearch Service Domains](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/off-peak.html)

* `enabled` - (Optional) Enabled disabled toggle for off-peak update window.
* `off_peak_window` - (Optional)
    * `window_start_time` - (Optional) 10h window for updates
        * `hours` - (Required) Starting hour of the 10-hour window for updates
        * `minutes` - (Required) Starting minute of the 10-hour window for updates

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the domain.
* `domain_id` - Unique identifier for the domain.
* `domain_name` - Name of the OpenSearch domain.
* `endpoint` - Domain-specific endpoint used to submit index, search, and data upload requests.
* `dashboard_endpoint` - Domain-specific endpoint for Dashboard without https scheme.
* `kibana_endpoint` - (**Deprecated**) Domain-specific endpoint for kibana without https scheme. Use the `dashboard_endpoint` attribute instead.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_options.0.availability_zones` - If the domain was created inside a VPC, the names of the availability zones the configured `subnet_ids` were created inside.
* `vpc_options.0.vpc_id` - If the domain was created inside a VPC, the ID of the VPC.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch domains using the `domain_name`. For example:

```terraform
import {
  to = aws_opensearch_domain.example
  id = "domain_name"
}
```

Using `terraform import`, import OpenSearch domains using the `domain_name`. For example:

```console
% terraform import aws_opensearch_domain.example domain_name
```
