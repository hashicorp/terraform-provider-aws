---
subcategory: "CloudHSM"
layout: "aws"
page_title: "AWS: aws_cloudhsm_v2_hsm"
description: |-
  Provides a CloudHSM v2 HSM module resource.
---

# Resource: aws_cloudhsm_v2_hsm

Creates an HSM module in Amazon CloudHSM v2 cluster.

## Example Usage

### Basic Usage

The following example below creates an HSM module in CloudHSM cluster.

```terraform
data "aws_cloudhsm_v2_cluster" "cluster" {
  cluster_id = var.cloudhsm_cluster_id
}

resource "aws_cloudhsm_v2_hsm" "cloudhsm_v2_hsm" {
  subnet_id  = data.aws_cloudhsm_v2_cluster.cluster.subnet_ids[0]
  cluster_id = data.aws_cloudhsm_v2_cluster.cluster.cluster_id
}
```

### Cluster Certificate Signing Request

A cluster's certificate signing request (CSR) is only available after the first HSM is added to the cluster. This example creates a cluster and its first HSM, then reads the CSR with the [`aws_cloudhsm_v2_cluster` data source](/docs/providers/aws/d/cloudhsm_v2_cluster.html). Sign the CSR and initialize the cluster outside of Terraform.

```terraform
resource "aws_cloudhsm_v2_cluster" "example" {
  hsm_type   = "hsm2m.medium"
  mode       = "FIPS"
  subnet_ids = [aws_subnet.example1.id, aws_subnet.example2.id]
}

resource "aws_cloudhsm_v2_hsm" "example" {
  cluster_id = aws_cloudhsm_v2_cluster.example.cluster_id
  subnet_id  = aws_subnet.example1.id
}

data "aws_cloudhsm_v2_cluster" "example" {
  cluster_id = aws_cloudhsm_v2_cluster.example.cluster_id

  depends_on = [aws_cloudhsm_v2_hsm.example]
}

output "cluster_csr" {
  value = data.aws_cloudhsm_v2_cluster.example.cluster_certificates[0].cluster_csr
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_id` - (Required) The ID of Cloud HSM v2 cluster to which HSM will be added.
* `subnet_id` - (Optional) The ID of subnet in which HSM module will be located. Conflicts with `availability_zone`.
* `availability_zone` - (Optional) The IDs of AZ in which HSM module will be located. Conflicts with `subnet_id`.
* `ip_address` - (Optional) The IP address of HSM module. Must be within the CIDR of selected subnet.

~> **NOTE:** Either `subnet_id` or `availability_zone` must be specified.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `availability_zone` - Name of the Availability Zone the HSM instance is located in.
* `cluster_id` - ID of Cloud HSM v2 cluster.
* `hsm_eni_id` - The ID of the ENI interface allocated for HSM module.
* `hsm_id` - The ID of the HSM module.
* `hsm_state` - The state of the HSM module.
* `ip_address` - The IP address of the HSM Module.
* `subnet_id` -  The ID of subnet in which HSM is located

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `120m`)
* `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import HSM modules using their HSM ID. For example:

```terraform
import {
  to = aws_cloudhsm_v2_hsm.bar
  id = "hsm-quo8dahtaca"
}
```

Using `terraform import`, import HSM modules using their HSM ID. For example:

```console
% terraform import aws_cloudhsm_v2_hsm.bar hsm-quo8dahtaca
```
