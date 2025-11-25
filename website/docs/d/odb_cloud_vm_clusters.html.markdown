---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_cloud_vm_clusters"
page_title: "AWS: aws_odb_cloud_vm_clusters"
description: |-
  Terraform data source for retrieving all cloud vm clusters resource in AWS for Oracle Database@AWS.
---

# Data Source: aws_odb_cloud_vm_clusters

Terraform data source for retrieving all cloud vm clusters AWS for Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_cloud_vm_clusters" "example" {}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cloud_vm_clusters` - List of Cloud VM Clusters. It returns only basic information about the cloud VM clusters.

### cloud_vm_clusters

* `id` - The unique identifier of the cloud vm cluster.
* `arn` - The Amazon Resource Name (ARN) for the cloud vm cluster.
* `cloud_exadata_infrastructure_id` - The ID of the Cloud Exadata Infrastructure.
* `oci_resource_anchor_name` - The name of the OCI Resource Anchor.
* `odb_network_id` - The ID of the ODB network.
* `oci_url` - The HTTPS link to the VM cluster in OCI.
* `ocid` - The OCID of the VM cluster.
* `display_name` - The display name of the VM cluster.
