---
subcategory: "Route53 Recovery Control Config"
layout: "aws"
page_title: "AWS: aws_route53recoverycontrolconfig_cluster"
description: |-
  Provides an AWS Route 53 Recovery Control Config Cluster
---

# Resource: aws_route53recoverycontrolconfig_cluster

Provides an AWS Route 53 Recovery Control Config Cluster

## Example Usage

```terraform
resource "aws_route53recoverycontrolconfig_cluster" "mycluster" {
  name = "mycluster"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name describing the cluster

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cluster_arn` - The ARN of the cluster
* `cluster_endpoints` - A list of 5 endpoints in 5 regions that can be used to talk to cluster
* `status` - Represents status of cluster. PENDING when its being created, PENDING_DELETION when its being  deleted and DEPLOYED otherwise

## Import

Route53 Recovery Control Config cluster can be imported via the cluster arn, e.g.

```
$ terraform import aws_route53recoverycontrolconfig_cluster.mycluster mycluster
```

## Timeouts

`aws_route53recoverycontrolconfig_cluster` has a timeout of 1 minute for creation and deletion