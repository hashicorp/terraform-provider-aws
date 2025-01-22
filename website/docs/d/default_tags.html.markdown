---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_default_tags"
description: |-
  Access the default tags configured on the provider.
---

# Data Source: aws_default_tags

Use this data source to get the default tags configured on the provider.

With this data source, you can apply default tags to resources not _directly_ managed by a Terraform resource, such as the instances underneath an Auto Scaling group or the volumes created for an EC2 instance.

## Example Usage

### Basic Usage

```terraform
data "aws_default_tags" "example" {}
```

### Dynamically Apply Default Tags to Auto Scaling Group

```terraform
provider "aws" {
  default_tags {
    tags = {
      Environment = "Test"
      Name        = "Provider Tag"
    }
  }
}

data "aws_default_tags" "example" {}

resource "aws_autoscaling_group" "example" {
  # ...
  dynamic "tag" {
    for_each = data.aws_default_tags.example.tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `tags` - Key-value mapping of provider default tags.
