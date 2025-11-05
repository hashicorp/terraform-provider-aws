---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_cloud_autonomous_vm_clusters"
page_title: "AWS: aws_odb_cloud_autonomous_vm_clusters"
description: |-
  Terraform data source for managing cloud autonomous vm clusters in AWS for Oracle Database@AWS.
---

# Data Source: aws_odb_cloud_autonomous_vm_clusters

Terraform data source for managing cloud autonomous vm clusters in AWS for Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_cloud_autonomous_vm_clusters" "example" {}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cloud_autonomous_vm_clusters` - List of Cloud Autonomous VM Clusters. The list going to contain basic information about the cloud autonomous VM clusters.

### cloud_autonomous_vm_clusters

* `id` - The unique identifier of the cloud autonomous vm cluster.
* `arn` - The Amazon Resource Name (ARN) for the Exadata infrastructure.
* `cloud_exadata_infrastructure_id` - Cloud exadata infrastructure id associated with this cloud autonomous VM cluster.
* `odb_network_id` - The unique identifier of the ODB network associated with this Autonomous VM cluster.
* `oci_resource_anchor_name` - The name of the OCI resource anchor associated with this Autonomous VM cluster.
* `oci_url` - The URL for accessing the OCI console page for this Autonomous VM cluster.
* `ocid` - The Oracle Cloud Identifier (OCID) of the Autonomous VM cluster.
* `display_name` - The display name of the Autonomous VM cluster.
