---
subcategory: "Route 53 Recovery Control Config"
layout: "aws"
page_title: "AWS: aws_route53recoverycontrolconfig_cluster"
description: |-
  Provides an AWS Route 53 Recovery Control Config Cluster
---

# Resource: aws_route53recoverycontrolconfig_cluster

Provides an AWS Route 53 Recovery Control Config Cluster.

## Example Usage

```terraform
resource "aws_route53recoverycontrolconfig_cluster" "example" {
  name = "georgefitzgerald"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Unique name describing the cluster.
* `network_type` - (Optional) Network type of cluster. Valid values are `IPV4` and `DUALSTACK`. Defaults to `IPV4`.

The following arguments are optional:

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cluster
* `cluster_endpoints` - List of 5 endpoints in 5 regions that can be used to talk to the cluster. See below.
* `status` - Status of cluster. `PENDING` when it is being created, `PENDING_DELETION` when it is being deleted and `DEPLOYED` otherwise.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### cluster_endpoints

* `endpoint` - Cluster endpoint.
* `region` - Region of the endpoint.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Recovery Control Config cluster using the cluster ARN. For example:

```terraform
import {
  to = aws_route53recoverycontrolconfig_cluster.mycluster
  id = "arn:aws:route53-recovery-control::313517334327:cluster/f9ae13be-a11e-4ec7-8522-94a70468e6ea"
}
```

Using `terraform import`, import Route53 Recovery Control Config cluster using the cluster ARN. For example:

```console
% terraform import aws_route53recoverycontrolconfig_cluster.mycluster arn:aws:route53-recovery-control::313517334327:cluster/f9ae13be-a11e-4ec7-8522-94a70468e6ea
```
