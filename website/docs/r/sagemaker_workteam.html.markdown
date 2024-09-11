---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_workteam"
description: |-
  Provides a SageMaker Workteam resource.
---

# Resource: aws_sagemaker_workteam

Provides a SageMaker Workteam resource.

## Example Usage

### Cognito Usage

```terraform
resource "aws_sagemaker_workteam" "example" {
  workteam_name  = "example"
  workforce_name = aws_sagemaker_workforce.example.id
  description    = "example"

  member_definition {
    cognito_member_definition {
      client_id  = aws_cognito_user_pool_client.example.id
      user_pool  = aws_cognito_user_pool_domain.example.user_pool_id
      user_group = aws_cognito_user_group.example.id
    }
  }
}
```

### Oidc Usage

```terraform
resource "aws_sagemaker_workteam" "example" {
  workteam_name  = "example"
  workforce_name = aws_sagemaker_workforce.example.id
  description    = "example"

  member_definition {
    oidc_member_definition {
      groups = ["example"]
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Required) A description of the work team.
* `workforce_name` - (Required) The name of the Workteam (must be unique).
* `workteam_name` - (Required) The name of the workforce.
* `member_definition` - (Required) A list of Member Definitions that contains objects that identify the workers that make up the work team. Workforces can be created using Amazon Cognito or your own OIDC Identity Provider (IdP). For private workforces created using Amazon Cognito use `cognito_member_definition`. For workforces created using your own OIDC identity provider (IdP) use `oidc_member_definition`. Do not provide input for both of these parameters in a single request. see [Member Definition](#member-definition) details below.
* `notification_configuration` - (Optional) Configures notification of workers regarding available or expiring work items. see [Notification Configuration](#notification-configuration) details below.
* `worker_access_configuration` - (Optional) Use this optional parameter to constrain access to an Amazon S3 resource based on the IP address using supported IAM global condition keys. The Amazon S3 resource is accessed in the worker portal using a Amazon S3 presigned URL. see [Worker Access Configuration](#worker-access-configuration) details below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Member Definition

* `cognito_member_definition` - (Optional) The Amazon Cognito user group that is part of the work team. See [Cognito Member Definition](#cognito-member-definition) details below.
* `oidc_member_definition` - (Optional) A list user groups that exist in your OIDC Identity Provider (IdP). One to ten groups can be used to create a single private work team. See [Cognito Member Definition](#oidc-member-definition) details below.

#### Cognito Member Definition

* `client_id` - (Required) An identifier for an application client. You must create the app client ID using Amazon Cognito.
* `user_pool` - (Required) An identifier for a user pool. The user pool must be in the same region as the service that you are calling.
* `user_group` - (Required) An identifier for a user group.

#### Oidc Member Definition

* `groups` - (Required) A list of comma separated strings that identifies user groups in your OIDC IdP. Each user group is made up of a group of private workers.

### Notification Configuration

* `notification_topic_arn` - (Required) The ARN for the SNS topic to which notifications should be published.

### Worker Access Configuration

* `s3_presign` - (Required) Defines any Amazon S3 resource constraints. see [S3 Presign](#s3-presign) details below.

#### S3 Presign

* `iam_policy_constraints` - (Required) Use this parameter to specify the allowed request source. Possible sources are either SourceIp or VpcSourceIp. see [IAM Policy Constraints](#iam-policy-constraints) details below.

##### IAM Policy Constraints

* `source_ip` - (Optional) When SourceIp is Enabled the worker's IP address when a task is rendered in the worker portal is added to the IAM policy as a Condition used to generate the Amazon S3 presigned URL. This IP address is checked by Amazon S3 and must match in order for the Amazon S3 resource to be rendered in the worker portal. Valid values are `Enabled` or `Disabled`
* `vpc_source_ip` - (Optional) When VpcSourceIp is Enabled the worker's IP address when a task is rendered in private worker portal inside the VPC is added to the IAM policy as a Condition used to generate the Amazon S3 presigned URL. To render the task successfully Amazon S3 checks that the presigned URL is being accessed over an Amazon S3 VPC Endpoint, and that the worker's IP address matches the IP address in the IAM policy. To learn more about configuring private worker portal, see [Use Amazon VPC mode from a private worker portal](https://docs.aws.amazon.com/sagemaker/latest/dg/samurai-vpc-worker-portal.html). Valid values are `Enabled` or `Disabled`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Workteam.
* `id` - The name of the Workteam.
* `subdomain` - The subdomain for your OIDC Identity Provider.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Workteams using the `workteam_name`. For example:

```terraform
import {
  to = aws_sagemaker_workteam.example
  id = "example"
}
```

Using `terraform import`, import SageMaker Workteams using the `workteam_name`. For example:

```console
% terraform import aws_sagemaker_workteam.example example
```
