---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_group_tag"
description: |-
  Manages an individual Autoscaling Group tag
---

# Resource: aws_autoscaling_group_tag

Manages an individual Autoscaling Group (ASG) tag. This resource should only be used in cases where ASGs are created outside Terraform (e.g., ASGs implicitly created by EKS Node Groups).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_autoscaling_group` and `aws_autoscaling_group_tag` to manage tags of the same ASG will cause a perpetual difference where the `aws_autoscaling_group` resource will try to remove the tag being added by the `aws_autoscaling_group_tag` resource.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
resource "aws_eks_node_group" "example" {
  cluster_name    = "example"
  node_group_name = "example"

  # ... other configuration ...
}

resource "aws_autoscaling_group_tag" "example" {
  for_each = toset(
    [for asg in flatten(
      [for resources in aws_eks_node_group.example.resources : resources.autoscaling_groups]
    ) : asg.name]
  )

  autoscaling_group_name = each.value

  tag {
    key   = "k8s.io/cluster-autoscaler/node-template/label/eks.amazonaws.com/capacityType"
    value = "SPOT"

    propagate_at_launch = false
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `autoscaling_group_name` - (Required) Name of the Autoscaling Group to apply the tag to.
* `tag` - (Required) Tag to create. The `tag` block is documented below.

The `tag` block supports the following arguments:

* `key` - (Required) Tag name.
* `value` - (Required) Tag value.
* `propagate_at_launch` - (Required) Whether to propagate the tags to instances launched by the ASG.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ASG name and key, separated by a comma (`,`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_autoscaling_group_tag` using the ASG name and key, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_autoscaling_group_tag.example
  id = "asg-example,k8s.io/cluster-autoscaler/node-template/label/eks.amazonaws.com/capacityType"
}
```

Using `terraform import`, import `aws_autoscaling_group_tag` using the ASG name and key, separated by a comma (`,`). For example:

```console
% terraform import aws_autoscaling_group_tag.example asg-example,k8s.io/cluster-autoscaler/node-template/label/eks.amazonaws.com/capacityType
```
