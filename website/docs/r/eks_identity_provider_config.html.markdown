---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_identity_provider_config"
description: |-
  Manages an EKS Identity Provider Configuration.
---

# Resource: aws_eks_identity_provider_config

Manages an EKS Identity Provider Configuration.

## Example Usage

```terraform
resource "aws_eks_identity_provider_config" "example" {
  cluster_name = aws_eks_cluster.example.name

  oidc {
    client_id                     = "your client_id"
    identity_provider_config_name = "example"
    issuer_url                    = "your issuer_url"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_name` - (Required) Name of the EKS Cluster.
* `oidc` - (Required) Nested attribute containing [OpenID Connect](https://openid.net/connect/) identity provider information for the cluster. Detailed below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### oidc Configuration Block

* `client_id` - (Required) Client ID for the OpenID Connect identity provider.
* `groups_claim` - (Optional) The JWT claim that the provider will use to return groups.
* `groups_prefix` - (Optional) A prefix that is prepended to group claims e.g., `oidc:`.
* `identity_provider_config_name` - (Required) The name of the identity provider config.
* `issuer_url` - (Required) Issuer URL for the OpenID Connect identity provider.
* `required_claims` - (Optional) The key value pairs that describe required claims in the identity token.
* `username_claim` - (Optional) The JWT claim that the provider will use as the username.
* `username_prefix` - (Optional) A prefix that is prepended to username claims.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the EKS Identity Provider Configuration.
* `id` - EKS Cluster name and EKS Identity Provider Configuration name separated by a colon (`:`).
* `status` - Status of the EKS Identity Provider Configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `40m`)
* `delete` - (Default `40m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EKS Identity Provider Configurations using the `cluster_name` and `identity_provider_config_name` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_eks_identity_provider_config.my_identity_provider_config
  id = "my_cluster:my_identity_provider_config"
}
```

Using `terraform import`, import EKS Identity Provider Configurations using the `cluster_name` and `identity_provider_config_name` separated by a colon (`:`). For example:

```console
% terraform import aws_eks_identity_provider_config.my_identity_provider_config my_cluster:my_identity_provider_config
```
