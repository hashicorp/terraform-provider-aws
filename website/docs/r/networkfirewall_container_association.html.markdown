---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_container_association"
description: |-
  Manages an AWS Network Firewall Container Association.
---

# Resource: aws_networkfirewall_container_association

Manages an AWS Network Firewall Container Association. A container association links Amazon ECS or Amazon EKS clusters to Network Firewall, resolving container IP addresses into a dynamic IP set you can reference from stateful rule groups.

## Example Usage

### EKS Cluster

```terraform
resource "aws_networkfirewall_container_association" "example" {
  container_association_name = "example-eks-association"
  type                       = "EKS"
  description                = "Association for production EKS cluster"

  container_monitoring_configurations {
    cluster_arn = aws_eks_cluster.example.arn

    attribute_filters {
      key   = "app"
      value = "backend"
    }
  }

  tags = {
    Name        = "example"
    Environment = "production"
  }
}
```

### ECS Cluster

```terraform
resource "aws_networkfirewall_container_association" "example" {
  container_association_name = "example-ecs-association"
  type                       = "ECS"

  container_monitoring_configurations {
    cluster_arn = aws_ecs_cluster.example.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `container_association_name` - (Required) Name of the container association. You can't change the name after creation. Must be between 1 and 128 characters and contain only alphanumeric characters and hyphens.
* `container_monitoring_configurations` - (Required) One or more monitoring configurations, up to 5. See [`container_monitoring_configurations` Block](#container_monitoring_configurations-block) below.
* `description` - (Optional) Description of the container association.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Required) Container orchestration platform for the clusters in this association. Valid values: `ECS`, `EKS`. You can't change the type after creation.

### `container_monitoring_configurations` Block

The `container_monitoring_configurations` block supports the following arguments:

* `attribute_filters` - (Optional) Key-value pairs that filter which containers within the cluster are monitored. For Amazon EKS, filter by namespace and Kubernetes labels. For Amazon ECS, filter by container instance attributes; attribute filters only match containers on the EC2 launch type, not Fargate. See [`attribute_filters` Block](#attribute_filters-block) below.
* `cluster_arn` - (Required) ARN of the Amazon ECS or Amazon EKS cluster to monitor. The cluster must be in the same Region and account as the container association.

### `attribute_filters` Block

The `attribute_filters` block supports the following arguments:

* `key` - (Required) Key of the container attribute to filter on.
* `value` - (Required) Value of the container attribute to filter on.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `container_association_arn` - ARN of the container association.
* `last_updated_time` - Date and time that the container association was last updated or resolved new container IP addresses.
* `resolved_cidr_count` - Number of CIDR blocks resolved from the monitored containers for this container association.
* `status` - Current status of the container association. Valid values: `CREATING`, `ACTIVE`, `UPDATING`, `DELETING`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_token` - Token used for optimistic locking.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networkfirewall_container_association.example
  identity = {
    "container_association_arn" = "arn:aws:network-firewall:us-west-2:123456789012:container-association/example"
  }
}

resource "aws_networkfirewall_container_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `container_association_arn` (String) ARN of the container association.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Container Association using the `container_association_arn`. For example:

```terraform
import {
  to = aws_networkfirewall_container_association.example
  id = "arn:aws:network-firewall:us-west-2:123456789012:container-association/example"
}
```

Using `terraform import`, import Network Firewall Container Association using the `container_association_arn`. For example:

```console
% terraform import aws_networkfirewall_container_association.example arn:aws:network-firewall:us-west-2:123456789012:container-association/example
```
