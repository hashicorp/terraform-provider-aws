---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_region"
description: |-
    Provides details about a specific AWS Region
---

# Data Source: aws_region

`aws_region` provides details about a specific AWS Region.

As well as validating a given Region name this resource can be used to
discover the name of the Region configured within the provider. The latter
can be useful in a child module which is inheriting an AWS provider
configuration from its parent module.

## Example Usage

The following example shows how the resource might be used to obtain
the name of the AWS Region configured on the provider.

```terraform
data "aws_region" "current" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Full name of the region to select (e.g. `us-east-1`), and the region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `endpoint` - (Optional) EC2 endpoint of the region to select.
* `name` - (Optional, **Deprecated**) Full name of the region to select. Use `region` instead.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Region's name (e.g. `us-east-1`).
* `description` - Region's description in this format: "Location (Region name)".
