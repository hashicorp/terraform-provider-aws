---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_parameter_group"
description: |-
  Provides a Redshift Cluster parameter group resource.
---

# Resource: aws_redshift_parameter_group

Provides a Redshift Cluster parameter group resource.

## Example Usage

```terraform
resource "aws_redshift_parameter_group" "bar" {
  name   = "parameter-group-test-terraform"
  family = "redshift-1.0"

  parameter {
    name  = "require_ssl"
    value = "true"
  }

  parameter {
    name  = "query_group"
    value = "example"
  }

  parameter {
    name  = "enable_user_activity_logging"
    value = "true"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the Redshift parameter group.
* `family` - (Required) The family of the Redshift parameter group.
* `description` - (Optional) The description of the Redshift parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of Redshift parameters to apply.

Parameter blocks support the following:

* `name` - (Required) The name of the Redshift parameter.
* `value` - (Required) The value of the Redshift parameter.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

You can read more about the parameters that Redshift supports in the [documentation](http://docs.aws.amazon.com/redshift/latest/mgmt/working-with-parameter-groups.html)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of parameter group
* `id` - The Redshift parameter group name.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Parameter Groups using the `name`. For example:

```terraform
import {
  to = aws_redshift_parameter_group.paramgroup1
  id = "parameter-group-test-terraform"
}
```

Using `terraform import`, import Redshift Parameter Groups using the `name`. For example:

```console
% terraform import aws_redshift_parameter_group.paramgroup1 parameter-group-test-terraform
```
