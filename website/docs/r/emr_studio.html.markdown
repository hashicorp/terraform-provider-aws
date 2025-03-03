---
subcategory: "EMR"
layout: "aws"
page_title: "AWS: aws_emr_studio"
description: |-
  Provides an Elastic MapReduce Studio
---

# Resource: aws_emr_studio

Provides an Elastic MapReduce Studio.

## Example Usage

```terraform
resource "aws_emr_studio" "example" {
  auth_mode                   = "SSO"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = "example"
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = [aws_subnet.test.id]
  user_role                   = aws_iam_role.test.arn
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id
}
```

## Argument Reference

The following arguments are required:

* `auth_mode`- (Required) Specifies whether the Studio authenticates users using IAM or Amazon Web Services SSO. Valid values are `SSO` or `IAM`.
* `default_s3_location` - (Required) The Amazon S3 location to back up Amazon EMR Studio Workspaces and notebook files.
* `name` - (Required) A descriptive name for the Amazon EMR Studio.
* `engine_security_group_id` - (Required) The ID of the Amazon EMR Studio Engine security group. The Engine security group allows inbound network traffic from the Workspace security group, and it must be in the same VPC specified by `vpc_id`.
* `service_role` - (Required) The IAM role that the Amazon EMR Studio assumes. The service role provides a way for Amazon EMR Studio to interoperate with other Amazon Web Services services.
* `subnet_ids` - (Required) A list of subnet IDs to associate with the Amazon EMR Studio. A Studio can have a maximum of 5 subnets. The subnets must belong to the VPC specified by `vpc_id`. Studio users can create a Workspace in any of the specified subnets.
* `vpc_id` - (Required) The ID of the Amazon Virtual Private Cloud (Amazon VPC) to associate with the Studio.
* `workspace_security_group_id` - (Required) The ID of the Amazon EMR Studio Workspace security group. The Workspace security group allows outbound network traffic to resources in the Engine security group, and it must be in the same VPC specified by `vpc_id`.

The following arguments are optional:

* `description` - (Optional) A detailed description of the Amazon EMR Studio.
* `encryption_key_arn` - (Optional) The AWS KMS key identifier (ARN) used to encrypt Amazon EMR Studio workspace and notebook files when backed up to Amazon S3.
* `idp_auth_url` - (Optional) The authentication endpoint of your identity provider (IdP). Specify this value when you use IAM authentication and want to let federated users log in to a Studio with the Studio URL and credentials from your IdP. Amazon EMR Studio redirects users to this endpoint to enter credentials.
* `idp_relay_state_parameter_name` - (Optional) The name that your identity provider (IdP) uses for its RelayState parameter. For example, RelayState or TargetSource. Specify this value when you use IAM authentication and want to let federated users log in to a Studio using the Studio URL. The RelayState parameter differs by IdP.
* `tags` - (Optional) list of tags to apply to the EMR Cluster. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_role` - (Optional) - The IAM user role that users and groups assume when logged in to an Amazon EMR Studio. Only specify a User Role when you use Amazon Web Services SSO authentication. The permissions attached to the User Role can be scoped down for each user or group using session policies.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn`- ARN of the studio.
* `url` - The unique access URL of the Amazon EMR Studio.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EMR studios using the `id`. For example:

```terraform
import {
  to = aws_emr_studio.studio
  id = "es-123456ABCDEF"
}
```

Using `terraform import`, import EMR studios using the `id`. For example:

```console
% terraform import aws_emr_studio.studio es-123456ABCDEF
```
