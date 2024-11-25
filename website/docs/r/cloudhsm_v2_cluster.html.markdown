---
subcategory: "CloudHSM"
layout: "aws"
page_title: "AWS: aws_cloudhsm_v2_cluster"
description: |-
  Provides a CloudHSM v2 resource.
---

# Resource: aws_cloudhsm_v2_cluster

Creates an Amazon CloudHSM v2 cluster.

For information about CloudHSM v2, see the
[AWS CloudHSM User Guide][1] and the [Amazon
CloudHSM API Reference][2].

~> **NOTE:** A CloudHSM Cluster can take several minutes to set up.
Practically no single attribute can be updated, except for `tags`.
If you need to delete a cluster, you have to remove its HSM modules first.
To initialize cluster, you have to add an HSM instance to the cluster, then sign CSR and upload it.

## Example Usage

The following example below creates a CloudHSM cluster.

```terraform
provider "aws" {
  region = var.aws_region
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "cloudhsm_v2_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}

resource "aws_subnet" "cloudhsm_v2_subnets" {
  count                   = 2
  vpc_id                  = aws_vpc.cloudhsm_v2_vpc.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = false
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}

resource "aws_cloudhsm_v2_cluster" "cloudhsm_v2_cluster" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.cloudhsm_v2_subnets[*].id

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `source_backup_identifier` - (Optional) ID of Cloud HSM v2 cluster backup to be restored.
* `hsm_type` - (Required) The type of HSM module in the cluster. Currently, `hsm1.medium` and `hsm2m.medium` are supported.
* `subnet_ids` - (Required) The IDs of subnets in which cluster will operate.
* `mode` - (Optional) The mode to use in the cluster. The allowed values are `FIPS` and `NON_FIPS`. This field is required if `hsm_type` is `hsm2m.medium`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `cluster_id` - The id of the CloudHSM cluster.
* `cluster_state` - The state of the CloudHSM cluster.
* `vpc_id` - The id of the VPC that the CloudHSM cluster resides in.
* `security_group_id` - The ID of the security group associated with the CloudHSM cluster.
* `cluster_certificates` - The list of cluster certificates.
    * `cluster_certificates.0.cluster_certificate` - The cluster certificate issued (signed) by the issuing certificate authority (CA) of the cluster's owner.
    * `cluster_certificates.0.cluster_csr` - The certificate signing request (CSR). Available only in `UNINITIALIZED` state after an HSM instance is added to the cluster.
    * `cluster_certificates.0.aws_hardware_certificate` - The HSM hardware certificate issued (signed) by AWS CloudHSM.
    * `cluster_certificates.0.hsm_certificate` - The HSM certificate issued (signed) by the HSM hardware.
    * `cluster_certificates.0.manufacturer_hardware_certificate` - The HSM hardware certificate issued (signed) by the hardware manufacturer.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

[1]: https://docs.aws.amazon.com/cloudhsm/latest/userguide/introduction.html
[2]: https://docs.aws.amazon.com/cloudhsm/latest/APIReference/Welcome.html

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudHSM v2 Clusters using the cluster `id`. For example:

```terraform
import {
  to = aws_cloudhsm_v2_cluster.test_cluster
  id = "cluster-aeb282a201"
}
```

Using `terraform import`, import CloudHSM v2 Clusters using the cluster `id`. For example:

```console
% terraform import aws_cloudhsm_v2_cluster.test_cluster cluster-aeb282a201
```
