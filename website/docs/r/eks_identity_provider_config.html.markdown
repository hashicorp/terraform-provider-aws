---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_identity_provider_config"
description: |-
  Manages an EKS Identity Provider Configuration
---

# Resource: aws_eks_identity_provider_config

Manages an EKS Identity Provider Configuration.

## Example Usage

```hcl
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

The following arguments are required:

* `cluster_name` – Name of the EKS Cluster.
* `oidc` - Nested attribute containing [OpenID Connect](https://openid.net/connect/) identity provider information for the cluster.
    * `client_id` – Client ID for the OpenID Connect identity provider.
    * `identity_provider_config_name` – The name of the identity provider config.
    * `issuer_url` - Issuer URL for the OpenID Connect identity provider.

The following arguments are optional:

* `oidc`
    * `groups_claim` - The JWT claim that the provider will use to return groups.
    * `groups_prefix` - A prefix that is prepended to group claims e.g. `oidc:`.
    * `required_claims` - The key value pairs that describe required claims in the identity token.
    * `username_claim` - The JWT claim that the provider will use as the username.
    * `username_prefix` - A prefix that is prepended to username claims.
* `tags` - (Optional) Key-value map of resource tags.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - EKS Cluster name and EKS Identity Provider Configuration name separated by a colon (`:`).
* `status` - Status of the EKS Identity Provider Configuration.

## Timeouts

`aws_eks_identity_provider_config` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `25 minutes`) How long to wait for the EKS Identity Provider Configuration to be associated.
* `delete` - (Default `25 minutes`) How long to wait for the EKS Identity Provider Configuration to be disassociated.

## Import

EKS Identity Provider Configuration can be imported using the `cluster_name` and `identity_provider_config_name` separated by a colon (`:`), e.g.

```
$ terraform import aws_eks_identity_provider_config.my_identity_provider_config my_cluster:my_identity_provider_config
```
