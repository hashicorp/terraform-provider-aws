---
subcategory: "EKS"
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

The following arguments are supported:

* `cluster_name` – (Required) Name of the EKS Cluster.
* `oidc` - (Required) Nested attribute containing [OpenID Connect](https://openid.net/connect/) identity provider information for the cluster. Detailed below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### oidc Configuration Block

* `client_id` – (Required) Client ID for the OpenID Connect identity provider.
* `groups_claim` - (Optional) The JWT claim that the provider will use to return groups.
* `groups_prefix` - (Optional) A prefix that is prepended to group claims e.g., `oidc:`.
* `identity_provider_config_name` – (Required) The name of the identity provider config.
* `issuer_url` - (Required) Issuer URL for the OpenID Connect identity provider.
* `required_claims` - (Optional) The key value pairs that describe required claims in the identity token.
* `username_claim` - (Optional) The JWT claim that the provider will use as the username.
* `username_prefix` - (Optional) A prefix that is prepended to username claims.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EKS Identity Provider Configuration.
* `id` - EKS Cluster name and EKS Identity Provider Configuration name separated by a colon (`:`).
* `status` - Status of the EKS Identity Provider Configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_eks_identity_provider_config` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `40 minutes`) How long to wait for the EKS Identity Provider Configuration to be associated.
* `delete` - (Default `40 minutes`) How long to wait for the EKS Identity Provider Configuration to be disassociated.

## Import

EKS Identity Provider Configurations can be imported using the `cluster_name` and `identity_provider_config_name` separated by a colon (`:`), e.g.,

```
$ terraform import aws_eks_identity_provider_config.my_identity_provider_config my_cluster:my_identity_provider_config
```
