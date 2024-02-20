---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_cluster"
description: |-
  Retrieve information about an EKS Cluster
---

# Data Source: aws_eks_cluster

Retrieve information about an EKS Cluster.

## Example Usage

```terraform
data "aws_eks_cluster" "example" {
  name = "example"
}

output "endpoint" {
  value = data.aws_eks_cluster.example.endpoint
}

output "kubeconfig-certificate-authority-data" {
  value = data.aws_eks_cluster.example.certificate_authority[0].data
}

# Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019.
output "identity-oidc-issuer" {
  value = data.aws_eks_cluster.example.identity[0].oidc[0].issuer
}
```

## Argument Reference

* `name` - (Required) Name of the cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the cluster
* `arn` - ARN of the cluster.
* `access_config` - Configuration block for access config.
    * `authentication_mode` - Values returned are `CONFIG_MAP`, `API` or `API_AND_CONFIG_MAP`
    * `bootstrap_cluster_creator_admin_permissions` - Default to `true`.
* `certificate_authority` - Nested attribute containing `certificate-authority-data` for your cluster.
    * `data` - The base64 encoded certificate data required to communicate with your cluster. Add this to the `certificate-authority-data` section of the `kubeconfig` file for your cluster.
* `cluster_id` - The ID of your local Amazon EKS cluster on the AWS Outpost. This attribute isn't available for an AWS EKS cluster on AWS cloud.
* `created_at` - Unix epoch time stamp in seconds for when the cluster was created.
* `enabled_cluster_log_types` - The enabled control plane logs.
* `endpoint` - Endpoint for your Kubernetes API server.
* `identity` - Nested attribute containing identity provider information for your cluster. Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019. For an example using this information to enable IAM Roles for Service Accounts, see the [`aws_eks_cluster` resource documentation](/docs/providers/aws/r/eks_cluster.html).
    * `oidc` - Nested attribute containing [OpenID Connect](https://openid.net/connect/) identity provider information for the cluster.
        * `issuer` - Issuer URL for the OpenID Connect identity provider.
* `kubernetes_network_config` - Nested list containing Kubernetes Network Configuration.
    * `ip_family` - `ipv4` or `ipv6`.
    * `service_ipv4_cidr` - The CIDR block to assign Kubernetes pod and service IP addresses from if `ipv4` was specified when the cluster was created.
    * `service_ipv6_cidr` - The CIDR block to assign Kubernetes pod and service IP addresses from if `ipv6` was specified when the cluster was created. Kubernetes assigns service addresses from the unique local address range (fc00::/7) because you can't specify a custom IPv6 CIDR block when you create the cluster.
* `outpost_config` - Contains Outpost Configuration.
    * `control_plane_instance_type` - The Amazon EC2 instance type for all Kubernetes control plane instances.
    * `control_plane_placement` - An object representing the placement configuration for all the control plane instances of your local Amazon EKS cluster on AWS Outpost.
        * `group_name` - The name of the placement group for the Kubernetes control plane instances.
    * `outpost_arns` - List of ARNs of the Outposts hosting the EKS cluster. Only a single ARN is supported currently.
* `platform_version` - Platform version for the cluster.
* `role_arn` - ARN of the IAM role that provides permissions for the Kubernetes control plane to make calls to AWS API operations on your behalf.
* `status` - Status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`.
* `tags` - Key-value map of resource tags.
* `version` - Kubernetes server version for the cluster.
* `vpc_config` - Nested list containing VPC configuration for the cluster.
    * `cluster_security_group_id` - The cluster security group that was created by Amazon EKS for the cluster.
    * `endpoint_private_access` - Indicates whether or not the Amazon EKS private API server endpoint is enabled.
    * `endpoint_public_access` - Indicates whether or not the Amazon EKS public API server endpoint is enabled.
    * `public_access_cidrs` - List of CIDR blocks. Indicates which CIDR blocks can access the Amazon EKS public API server endpoint.
    * `security_group_ids` – List of security group IDs
    * `subnet_ids` – List of subnet IDs
    * `vpc_id` – The VPC associated with your cluster.
