---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_cluster_registration"
description: |-
  Connects a Kubernetes cluster to the Amazon EKS control plane.
---

# Resource: aws_eks_cluster_registration

Connects a Kubernetes cluster to the Amazon EKS control plane.

## Example Usage

```terraform
resource "aws_eks_cluster_registration" "example" {
  name = "example"

  connector_config {
    provider = "EKS_ANYWHERE"
    role_arn = aws_iam_role.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `name` – (Required) Name of the cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\-_]+$`).
* `connector_config` - (Required) The configuration settings required to connect the Kubernetes cluster to the Amazon EKS control plane.

The following arguments are optional:

* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### connector_config

The following arguments are supported in the `connector_config` configuration block:

* `provider` - (Required) The cloud provider for the target cluster to connect.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the role that is authorized to request the connector configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the cluster.
* `connector_config` - (Required) The configuration settings required to connect the Kubernetes cluster to the Amazon EKS control plane.
* `created_at` - Unix epoch timestamp in seconds for when the cluster was created.
* `status` - Status of the EKS cluster. One of `ACTIVE`, `FAILED`, `PENDING`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

### connector_config

* `activation_id` - A unique code associated with the cluster for registration purposes.
* `activation_expiry` - The expiration time of the connected cluster. The cluster's YAML file must be applied through the native provider.

## Import

EKS Clusters can be imported using the `name`, e.g.,

```
$ terraform import aws_eks_cluster_registration.my_cluster my_cluster
```
