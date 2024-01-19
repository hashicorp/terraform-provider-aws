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

## Argument Reference

This resource supports the following arguments:

~> **NOTE:** Either `subnet_id` or `availability_zone` must be specified.

* `cluster_id` - (Required) The ID of Cloud HSM v2 cluster to which HSM will be added.
* `subnet_id` - (Optional) The ID of subnet in which HSM module will be located. Conflicts with `availability_zone`.
* `availability_zone` - (Optional) The IDs of AZ in which HSM module will be located. Conflicts with `subnet_id`.
* `ip_address` - (Optional) The IP address of HSM module. Must be within the CIDR of selected subnet.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `availability_zone` - Name of the Availability Zone the HSM instance is located in.
* `cluster_id` - ID of Cloud HSM v2 cluster.
* `hsm_eni_id` - The id of the ENI interface allocated for HSM module.
* `hsm_id` - The id of the HSM module.
* `hsm_state` - The state of the HSM module.
* `ip_address` - The IP address of the HSM Module.
* `subnet_id` -  The ID of subnet in which HSM is located

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
