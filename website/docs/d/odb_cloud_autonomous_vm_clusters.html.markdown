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
