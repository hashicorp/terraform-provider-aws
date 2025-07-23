---
subcategory: "CloudHSM"
layout: "aws"
page_title: "AWS: aws_cloudhsm_v2_cluster"
description: |-
  Get information on a CloudHSM v2 cluster.
---

# Data Source: aws_cloudhsm_v2_cluster

Use this data source to get information about a CloudHSM v2 cluster

## Example Usage

```terraform
data "aws_cloudhsm_v2_cluster" "cluster" {
  cluster_id = "cluster-testclusterid"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_id` - (Required) ID of Cloud HSM v2 cluster.
* `cluster_state` - (Optional) State of the cluster to be found.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `vpc_id` - ID of the VPC that the CloudHSM cluster resides in.
* `security_group_id` - ID of the security group associated with the CloudHSM cluster.
* `subnet_ids` - IDs of subnets in which cluster operates.
* `cluster_certificates` - The list of cluster certificates.
    * `cluster_certificates.0.cluster_certificate` - The cluster certificate issued (signed) by the issuing certificate authority (CA) of the cluster's owner.
    * `cluster_certificates.0.cluster_csr` - The certificate signing request (CSR). Available only in UNINITIALIZED state.
    * `cluster_certificates.0.aws_hardware_certificate` - The HSM hardware certificate issued (signed) by AWS CloudHSM.
    * `cluster_certificates.0.hsm_certificate` - The HSM certificate issued (signed) by the HSM hardware.
    * `cluster_certificates.0.manufacturer_hardware_certificate` - The HSM hardware certificate issued (signed) by the hardware manufacturer.
The number of available cluster certificates may vary depending on state of the cluster.
