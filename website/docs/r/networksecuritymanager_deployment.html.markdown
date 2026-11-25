---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_deployment"
description: |-
  Manages an AWS Network Security Manager Deployment.
---

# Resource: aws_networksecuritymanager_deployment

Manages an AWS Network Security Manager Deployment. A deployment applies one or more policies to the accounts and resources selected by a scope.

Changes to a deployment are published unless `is_published` is `false`.

## Example Usage

### Basic Usage

```terraform
resource "aws_networksecuritymanager_deployment" "example" {
  name        = "example"
  policy_arns = [aws_networksecuritymanager_policy.example.arn]
  scope_arn   = aws_networksecuritymanager_scope.example.arn

  configuration {
    enable_cross_account_visibility = false
  }
}
```

## Argument Reference

The following arguments are required:

* `configuration` - (Required) Settings for the deployment. See [`configuration` Block](#configuration-block) below.
* `name` - (Required, Forces new resource) Name of the deployment.
* `policy_arns` - (Required) Set of ARNs of the policies to deploy. Specify 1 or 2 policies.
* `scope_arn` - (Required) ARN of the scope that selects the accounts and resources the policies apply to.

The following arguments are optional:

* `description` - (Optional) Description of the deployment.
* `is_published` - (Optional) Whether to publish the deployment. When `true`, the deployment is saved in the published (`ACTIVE`) state. When `false`, it is saved as a draft (`DRAFT`); for a published deployment, the draft is saved on top of the published version. Defaults to `true`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

~> **Note:** A policy or scope used by a deployment cannot be deleted. Configure [`create_before_destroy`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#create_before_destroy) on referenced policy and scope resources so that a replacement is associated with the deployment before the old resource is deleted.

### `configuration` Block

* `enable_cross_account_visibility` - (Required) Whether to enable cross-account visibility for the deployment.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the deployment.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_deployment.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:deployment:0123456789abcdef"
  }
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the deployment.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Deployment using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_deployment.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:deployment:0123456789abcdef"
}
```

Using `terraform import`, import Network Security Manager Deployment using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_deployment.example arn:aws:network-security-manager:us-east-1:123456789012:deployment:0123456789abcdef
```
