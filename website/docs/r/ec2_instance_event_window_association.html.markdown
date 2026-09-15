---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_event_window_association"
description: |-
  Provides an EC2 Instance Event Window Association resource.
---

# Resource: aws_ec2_instance_event_window_association

Provides an EC2 Instance Event Window Association resource. This associates a target (one or more instance IDs, instance tags, or Dedicated Host IDs) with an [EC2 Instance Event Window](ec2_instance_event_window.html.markdown).

## Example Usage

### Instance IDs

```terraform
resource "aws_ec2_instance_event_window_association" "example" {
  instance_event_window_id = aws_ec2_instance_event_window.example.id

  association_target {
    instance_ids = [aws_instance.example.id]
  }
}
```

### Instance Tags

```terraform
resource "aws_ec2_instance_event_window_association" "example" {
  instance_event_window_id = aws_ec2_instance_event_window.example.id

  association_target {
    instance_tags = {
      Environment = "production"
    }
  }
}
```

### Dedicated Host IDs

```terraform
resource "aws_ec2_instance_event_window_association" "example" {
  instance_event_window_id = aws_ec2_instance_event_window.example.id

  association_target {
    dedicated_host_ids = [aws_ec2_host.example.id]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `instance_event_window_id` - (Required) ID of the event window.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `association_target` - (Required) Configuration block for the association target. Only one target type can be associated with an event window at a time. See [`association_target`](#association_target) below.

### `association_target`

* `dedicated_host_ids` - (Optional) List of Dedicated Host IDs to associate with the event window. Conflicts with `instance_ids` and `instance_tags`. Exactly one of `dedicated_host_ids`, `instance_ids`, or `instance_tags` must be specified. Limit: 50.
* `instance_ids` - (Optional) List of instance IDs to associate with the event window. Conflicts with `dedicated_host_ids` and `instance_tags`. Exactly one of `dedicated_host_ids`, `instance_ids`, or `instance_tags` must be specified. Limit: 100.
* `instance_tags` - (Optional) Map of tag keys and values. Instances matching these tags are associated with the event window. Conflicts with `dedicated_host_ids` and `instance_ids`. Exactly one of `dedicated_host_ids`, `instance_ids`, or `instance_tags` must be specified. Limit: 50 tags. The instance tags can be associated with any number of instances.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the event window (same value as `instance_event_window_id`).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_instance_event_window_association.example
  identity = {
    id = "iew-0123456789abcdef0"
  }
}

resource "aws_ec2_instance_event_window_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the event window.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Instance Event Window Associations using the event window `id`. For example:

```terraform
import {
  to = aws_ec2_instance_event_window_association.example
  id = "iew-0123456789abcdef0"
}
```

Using `terraform import`, import EC2 Instance Event Window Associations using the event window `id`. For example:

```console
% terraform import aws_ec2_instance_event_window_association.example iew-0123456789abcdef0
```
