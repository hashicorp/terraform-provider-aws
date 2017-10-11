---
layout: "aws"
page_title: "AWS: aws_elasticsearch_domain"
sidebar_current: "docs-aws-datasource-elasticsearch-domain"
description: |-
  Get information on an ElasticSearch Domain resource.
---

# aws_elasticsearch_domain

Use this data source to get information about an ElasticSearch Domain

## Example Usage

```hcl
data "aws_elasticsearch_domain" "my_domain" {
  domain_name = "my-domain-name"
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` – (Required) Name of the domain.


## Attributes Reference

The following attributes are exported:

* `domain_id` – Unique identifier for the domain.
* `endpoint` – Domain-specific endpoint used to submit index, search, and data upload requests.
* `created` – Status of the creation of the domain.
* `deleted` – Status of the deletion of the domain.
* `access_policies` – The policy document attached to the domain.
* `processing` – Status of a configuration change in the domain.
* `elasticsearch_version` – ElasticSearch version for the domain.
* `arn` – The Amazon Resource Name (ARN) of the domain.
* `cluster_config` - Cluster configuration of the domain.
** `instance_type` - Instance type of data nodes in the cluster.
** `instance_count` - Number of instances in the cluster.
** `dedicated_master_enabled` - Indicates whether dedicated master nodes are enabled for the cluster.
** `dedicated_master_type` - Instance type of the dedicated master nodes in the cluster.
** `dedicated_master_count` - Number of dedicated master nodes in the cluster.
** `zone_awareness_enabled` - Indicates whether zone awareness is enabled.
* `ebs_options` - EBS Options for the instances in the domain.
** `ebs_enabled` - Whether EBS volumes are attached to data nodes in the domain.
** `volume_type` - The type of EBS volumes attached to data nodes.
** `volume_size` - The size of EBS volumes attached to data nodes (in GB).
** `iops` - The baseline input/output (I/O) performance of EBS volumes
	attached to data nodes.
* `snapshot_options` – Domain snapshot related options.
** `automated_snapshot_start_hour` - Hour during which the service takes an automated daily
	snapshot of the indices in the domain.
* `advanced_options` - Key-value string pairs to specify advanced configuration options.
* `tags` - The tags assigned to the domain.
