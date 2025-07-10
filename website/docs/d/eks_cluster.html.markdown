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
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the cluster
* `arn` - ARN of the cluster.
* `access_config` - Configuration block for access config.
    * `authentication_mode` - Values returned are `CONFIG_MAP`, `API` or `API_AND_CONFIG_MAP`
    * `bootstrap_cluster_creator_admin_permissions` - Default to `true`.
* `compute_config` - Nested attribute containing compute capability configuration for EKS Auto Mode enabled cluster.
    * `enabled` - Whether the EKS Auto Mode compute capability is enabled or not.
    * `node_pools` - List of node pools for the EKS Auto Mode compute capability.
    * `node_role_arn` - The ARN of the IAM Role EKS will assign to EC2 Managed Instances in your EKS Auto Mode cluster.
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
    * `elastic_load_balancing` - Contains Elastic Load Balancing configuration for EKS Auto Mode enabled cluster.
        * `enabled` - Indicates if the load balancing capability is enabled for EKS Auto Mode enabled cluster.
    * `ip_family` - `ipv4` or `ipv6`.
    * `service_ipv4_cidr` - The CIDR block to assign Kubernetes pod and service IP addresses from if `ipv4` was specified when the cluster was created.
    * `service_ipv6_cidr` - The CIDR block to assign Kubernetes pod and service IP addresses from if `ipv6` was specified when the cluster was created. Kubernetes assigns service addresses from the unique local address range (fc00::/7) because you can't specify a custom IPv6 CIDR block when you create the cluster.
* `outpost_config` - Contains Outpost Configuration.
    * `control_plane_instance_type` - The Amazon EC2 instance type for all Kubernetes control plane instances.
    * `control_plane_placement` - An object representing the placement configuration for all the control plane instances of your local Amazon EKS cluster on AWS Outpost.
        * `group_name` - The name of the placement group for the Kubernetes control plane instances.
    * `outpost_arns` - List of ARNs of the Outposts hosting the EKS cluster. Only a single ARN is supported currently.
* `platform_version` - Platform version for the cluster.
* `remote_network_config` - Contains remote network configuration for EKS Hybrid Nodes.
    * `remote_node_networks` - The networks that can contain hybrid nodes.
        * `cidrs` - List of network CIDRs that can contain hybrid nodes.
    * `remote_pod_networks` - The networks that can contain pods that run Kubernetes webhooks on hybrid nodes.
        * `cidrs` - List of network CIDRs that can contain pods that run Kubernetes webhooks on hybrid nodes.
* `role_arn` - ARN of the IAM role that provides permissions for the Kubernetes control plane to make calls to AWS API operations on your behalf.
* `status` - Status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`.
* `storage_config` - Contains storage configuration for EKS Auto Mode enabled cluster.
    * `block_storage` - Contains block storage configuration for EKS Auto Mode enabled cluster.
        * `enabled` - Indicates if the block storage capability is enabled for EKS Auto Mode enabled cluster.
* `tags` - Key-value map of resource tags.
* `upgrade_policy` - Configuration block for the support policy to use for the cluster.
    * `support_type` - Support type to use for the cluster.
* `version` - Kubernetes server version for the cluster.
* `vpc_config` - Nested list containing VPC configuration for the cluster.
    * `cluster_security_group_id` - The cluster security group that was created by Amazon EKS for the cluster.
    * `endpoint_private_access` - Indicates whether or not the Amazon EKS private API server endpoint is enabled.
    * `endpoint_public_access` - Indicates whether or not the Amazon EKS public API server endpoint is enabled.
    * `public_access_cidrs` - List of CIDR blocks. Indicates which CIDR blocks can access the Amazon EKS public API server endpoint.
    * `security_group_ids` - List of security group IDs
    * `subnet_ids` - List of subnet IDs
    * `vpc_id` - The VPC associated with your cluster.
* `zonal_shift_config` - Contains Zonal Shift Configuration.
    * `enabled` - Whether zonal shift is enabled.
