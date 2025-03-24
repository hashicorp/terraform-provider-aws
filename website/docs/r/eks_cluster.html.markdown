---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_cluster"
description: |-
  Manages an EKS Cluster
---

# Resource: aws_eks_cluster

Manages an EKS Cluster.

> **Hands-on:** For an example of `aws_eks_cluster` in use, follow the [Provision an EKS Cluster](https://learn.hashicorp.com/tutorials/terraform/eks) tutorial on HashiCorp Learn.

## Example Usage

### EKS Cluster

```terraform
resource "aws_eks_cluster" "example" {
  name = "example"

  access_config {
    authentication_mode = "API"
  }

  role_arn = aws_iam_role.example.arn
  version  = "1.31"

  vpc_config {
    subnet_ids = [
      aws_subnet.az1.id,
      aws_subnet.az2.id,
      aws_subnet.az3.id,
    ]
  }

  # Ensure that IAM Role permissions are created before and deleted
  # after EKS Cluster handling. Otherwise, EKS will not be able to
  # properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy,
  ]
}

resource "aws_iam_role" "cluster" {
  name = "eks-cluster-example"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sts:AssumeRole",
          "sts:TagSession"
        ]
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}
```

### EKS Cluster with EKS Auto Mode

~> **NOTE:** When using EKS Auto Mode `compute_config.enabled`, `kubernetes_network_config.elastic_load_balancing.enabled`, and `storage_config.block_storage.enabled` must *ALL be set to `true`. Likewise for disabling EKS Auto Mode, all three arguments must be set to `false`. Enabling EKS Auto Mode also requires that `bootstrap_self_managed_addons` is set to `false`.

```terraform
resource "aws_eks_cluster" "example" {
  name = "example"

  access_config {
    authentication_mode = "API"
  }

  role_arn = aws_iam_role.cluster.arn
  version  = "1.31"

  bootstrap_self_managed_addons = false

  compute_config {
    enabled       = true
    node_pools    = ["general-purpose"]
    node_role_arn = aws_iam_role.node.arn
  }

  kubernetes_network_config {
    elastic_load_balancing {
      enabled = true
    }
  }

  storage_config {
    block_storage {
      enabled = true
    }
  }

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = true

    subnet_ids = [
      aws_subnet.az1.id,
      aws_subnet.az2.id,
      aws_subnet.az3.id,
    ]
  }

  # Ensure that IAM Role permissions are created before and deleted
  # after EKS Cluster handling. Otherwise, EKS will not be able to
  # properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.cluster_AmazonEKSComputePolicy,
    aws_iam_role_policy_attachment.cluster_AmazonEKSBlockStoragePolicy,
    aws_iam_role_policy_attachment.cluster_AmazonEKSLoadBalancingPolicy,
    aws_iam_role_policy_attachment.cluster_AmazonEKSNetworkingPolicy,
  ]
}

resource "aws_iam_role" "node" {
  name = "eks-auto-node-example"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole"]
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "node_AmazonEKSWorkerNodeMinimalPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodeMinimalPolicy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node_AmazonEC2ContainerRegistryPullOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPullOnly"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role" "cluster" {
  name = "eks-cluster-example"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sts:AssumeRole",
          "sts:TagSession"
        ]
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSComputePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSComputePolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSBlockStoragePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSBlockStoragePolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSLoadBalancingPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSLoadBalancingPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSNetworkingPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSNetworkingPolicy"
  role       = aws_iam_role.cluster.name
}
```

### EKS Cluster with EKS Hybrid Nodes

```terraform
resource "aws_eks_cluster" "example" {
  name = "example"

  access_config {
    authentication_mode = "API"
  }

  role_arn = aws_iam_role.cluster.arn
  version  = "1.31"

  remote_network_config {
    remote_node_networks {
      cidrs = ["172.16.0.0/18"]
    }
    remote_pod_networks {
      cidrs = ["172.16.64.0/18"]
    }
  }

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = true

    subnet_ids = [
      aws_subnet.az1.id,
      aws_subnet.az2.id,
      aws_subnet.az3.id,
    ]
  }

  # Ensure that IAM Role permissions are created before and deleted
  # after EKS Cluster handling. Otherwise, EKS will not be able to
  # properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy,
  ]
}

resource "aws_iam_role" "cluster" {
  name = "eks-cluster-example"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sts:AssumeRole",
          "sts:TagSession"
        ]
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}
```

### Local EKS Cluster on AWS Outpost

[Creating a local Amazon EKS cluster on an AWS Outpost](https://docs.aws.amazon.com/eks/latest/userguide/create-cluster-outpost.html)

```terraform
resource "aws_eks_cluster" "example" {
  name = "example"

  access_config {
    authentication_mode = "CONFIG_MAP"
  }

  role_arn = aws_iam_role.example.arn
  version  = "1.31"

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = false

    subnet_ids = [
      aws_subnet.az1.id,
      aws_subnet.az2.id,
      aws_subnet.az3.id,
    ]
  }

  outpost_config {
    control_plane_instance_type = "m5.large"
    outpost_arns                = [data.aws_outposts_outpost.example.arn]
  }

  # Ensure that IAM Role permissions are created before and deleted
  # after EKS Cluster handling. Otherwise, EKS will not be able to
  # properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSLocalOutpostClusterPolicy,
  ]
}

data "aws_outposts_outpost" "example" {
  name = "example"
}

resource "aws_iam_role" "cluster" {
  name = "eks-cluster-example"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sts:AssumeRole",
          "sts:TagSession"
        ]
        Effect = "Allow"
        Principal = {
          Service = [
            "eks.amazonaws.com",
            "ec2.amazonaws.com"
          ]
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSLocalOutpostClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSLocalOutpostClusterPolicy"
  role       = aws_iam_role.cluster.name
}
```

## Argument Reference

The following arguments are required:

* `name` – (Required) Name of the cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\-_]*$`).
* `role_arn` - (Required) ARN of the IAM role that provides permissions for the Kubernetes control plane to make calls to AWS API operations on your behalf. Ensure the resource configuration includes explicit dependencies on the IAM Role permissions by adding [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html) if using the [`aws_iam_role_policy` resource](/docs/providers/aws/r/iam_role_policy.html) or [`aws_iam_role_policy_attachment` resource](/docs/providers/aws/r/iam_role_policy_attachment.html), otherwise EKS cannot delete EKS managed EC2 infrastructure such as Security Groups on EKS Cluster deletion.
* `vpc_config` - (Required) Configuration block for the VPC associated with your cluster. Amazon EKS VPC resources have specific requirements to work properly with Kubernetes. For more information, see [Cluster VPC Considerations](https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html) and [Cluster Security Group Considerations](https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html) in the Amazon EKS User Guide. Detailed below. Also contains attributes detailed in the Attributes section.

The following arguments are optional:

* `access_config` - (Optional) Configuration block for the access config associated with your cluster, see [Amazon EKS Access Entries](https://docs.aws.amazon.com/eks/latest/userguide/access-entries.html). [Detailed](#access_config) below.
* `bootstrap_self_managed_addons` - (Optional) Install default unmanaged add-ons, such as `aws-cni`, `kube-proxy`, and CoreDNS during cluster creation. If `false`, you must manually install desired add-ons. Changing this value will force a new cluster to be created. Defaults to `true`.
* `compute_config` - (Optional) Configuration block with compute configuration for EKS Auto Mode. [Detailed](#compute_config) below.
* `enabled_cluster_log_types` - (Optional) List of the desired control plane logging to enable. For more information, see [Amazon EKS Control Plane Logging](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html).
* `encryption_config` - (Optional) Configuration block with encryption configuration for the cluster. [Detailed](#encryption_config) below.
* `kubernetes_network_config` - (Optional) Configuration block with kubernetes network configuration for the cluster. [Detailed](#kubernetes_network_config) below. If removed, Terraform will only perform drift detection if a configuration value is provided.
* `outpost_config` - (Optional) Configuration block representing the configuration of your local Amazon EKS cluster on an AWS Outpost. This block isn't available for creating Amazon EKS clusters on the AWS cloud.
* `remote_network_config` - (Optional) Configuration block with remote network configuration for EKS Hybrid Nodes. [Detailed](#remote_network_config) below.
* `storage_config` - (Optional) Configuration block with storage configuration for EKS Auto Mode. [Detailed](#storage_config) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `upgrade_policy` - (Optional) Configuration block for the support policy to use for the cluster.  See [upgrade_policy](#upgrade_policy) for details.
* `version` – (Optional) Desired Kubernetes master version. If you do not specify a value, the latest available version at resource creation is used and no upgrades will occur except those automatically triggered by EKS. The value must be configured and increased to upgrade the version when desired. Downgrades are not supported by EKS.
* `zonal_shift_config` - (Optional) Configuration block with zonal shift configuration for the cluster. [Detailed](#zonal_shift_config) below.

### access_config

The `access_config` configuration block supports the following arguments:

* `authentication_mode` - (Optional) The authentication mode for the cluster. Valid values are `CONFIG_MAP`, `API` or `API_AND_CONFIG_MAP`
* `bootstrap_cluster_creator_admin_permissions` - (Optional) Whether or not to bootstrap the access config values to the cluster. Default is `false`.

### compute_config

The `compute_config` configuration block supports the following arguments:

* `enabled` - (Optional) Request to enable or disable the compute capability on your EKS Auto Mode cluster. If the compute capability is enabled, EKS Auto Mode will create and delete EC2 Managed Instances in your Amazon Web Services account.
* `node_pools` - (Optional) Configuration for node pools that defines the compute resources for your EKS Auto Mode cluster. Valid options are `general-purpose` and `system`.
* `node_role_arn` - (Optional) The ARN of the IAM Role EKS will assign to EC2 Managed Instances in your EKS Auto Mode cluster. This value cannot be changed after the compute capability of EKS Auto Mode is enabled..

### encryption_config

The `encryption_config` configuration block supports the following arguments:

* `provider` - (Required) Configuration block with provider for encryption. Detailed below.
* `resources` - (Required) List of strings with resources to be encrypted. Valid values: `secrets`.

#### provider

The `provider` configuration block supports the following arguments:

* `key_arn` - (Required) ARN of the Key Management Service (KMS) customer master key (CMK). The CMK must be symmetric, created in the same region as the cluster, and if the CMK was created in a different account, the user must have access to the CMK. For more information, see [Allowing Users in Other Accounts to Use a CMK in the AWS Key Management Service Developer Guide](https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-modifying-external-accounts.html).

### remote_network_config

The `remote_network_config` configuration block supports the following arguments:

* `remote_node_networks` - (Optional) Configuration block with remote node network configuration for EKS Hybrid Nodes. Detailed below.
* `remote_pod_networks` - (Optional) Configuration block with remote pod network configuration for EKS Hybrid Nodes. Detailed below.

#### remote_node_networks

The `remote_node_networks` configuration block supports the following arguments:

* `cidrs` - (Required) List of network CIDRs that can contain hybrid nodes.

#### remote_pod_networks

The `remote_pod_networks` configuration block supports the following arguments:

* `cidrs` - (Required) List of network CIDRs that can contain pods that run Kubernetes webhooks on hybrid nodes.

### vpc_config Arguments

* `cluster_security_group_id` - (Computed) Cluster security group that is created by Amazon EKS for the cluster. Managed node groups use this security group for control-plane-to-data-plane communication.
* `endpoint_private_access` - (Optional) Whether the Amazon EKS private API server endpoint is enabled. Default is `false`.
* `endpoint_public_access` - (Optional) Whether the Amazon EKS public API server endpoint is enabled. Default is `true`.
* `public_access_cidrs` - (Optional) List of CIDR blocks. Indicates which CIDR blocks can access the Amazon EKS public API server endpoint when enabled. EKS defaults this to a list with `0.0.0.0/0`. Terraform will only perform drift detection of its value when present in a configuration.
* `security_group_ids` – (Optional) List of security group IDs for the cross-account elastic network interfaces that Amazon EKS creates to use to allow communication between your worker nodes and the Kubernetes control plane.
* `subnet_ids` – (Required) List of subnet IDs. Must be in at least two different availability zones. Amazon EKS creates cross-account elastic network interfaces in these subnets to allow communication between your worker nodes and the Kubernetes control plane.
* `vpc_id` - (Computed) ID of the VPC associated with your cluster.

### kubernetes_network_config

The `kubernetes_network_config` configuration block supports the following arguments:

* `elastic_load_balancing` - (Optional) Configuration block with elastic load balancing configuration for the cluster. Detailed below.
* `service_ipv4_cidr` - (Optional) The CIDR block to assign Kubernetes pod and service IP addresses from. If you don't specify a block, Kubernetes assigns addresses from either the 10.100.0.0/16 or 172.20.0.0/16 CIDR blocks. We recommend that you specify a block that does not overlap with resources in other networks that are peered or connected to your VPC. You can only specify a custom CIDR block when you create a cluster, changing this value will force a new cluster to be created. The block must meet the following requirements:

    * Within one of the following private IP address blocks: 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16.

    * Doesn't overlap with any CIDR block assigned to the VPC that you selected for VPC.

    * Between /24 and /12.

* `service_ipv6_cidr` - (Computed) The CIDR block that Kubernetes pod and service IP addresses are assigned from if you specify `ipv6` for `ip_family` when you create the cluster. Kubernetes assigns service addresses from the unique local address range (fc00::/7) because you can't specify a custom IPv6 CIDR block when you create the cluster.
* `ip_family` - (Optional) The IP family used to assign Kubernetes pod and service addresses. Valid values are `ipv4` (default) and `ipv6`. You can only specify an IP family when you create a cluster, changing this value will force a new cluster to be created.

#### elastic_load_balancing

The `elastic_load_balancing` configuration block supports the following arguments:

* `enabled` - (Optional) Indicates if the load balancing capability is enabled on your EKS Auto Mode cluster. If the load balancing capability is enabled, EKS Auto Mode will create and delete load balancers in your Amazon Web Services account.

### outpost_config

The `outpost_config` configuration block supports the following arguments:

* `control_plane_instance_type` - (Required) The Amazon EC2 instance type that you want to use for your local Amazon EKS cluster on Outposts. The instance type that you specify is used for all Kubernetes control plane instances. The instance type can't be changed after cluster creation. Choose an instance type based on the number of nodes that your cluster will have. If your cluster will have:

    * 1–20 nodes, then we recommend specifying a large instance type.

    * 21–100 nodes, then we recommend specifying an xlarge instance type.

    * 101–250 nodes, then we recommend specifying a 2xlarge instance type.

    For a list of the available Amazon EC2 instance types, see Compute and storage in AWS Outposts rack features  The control plane is not automatically scaled by Amazon EKS.

* `control_plane_placement` - (Optional) An object representing the placement configuration for all the control plane instances of your local Amazon EKS cluster on AWS Outpost.
The `control_plane_placement` configuration block supports the following arguments:

    * `group_name` - (Required) The name of the placement group for the Kubernetes control plane instances. This setting can't be changed after cluster creation.

* `outpost_arns` - (Required) The ARN of the Outpost that you want to use for your local Amazon EKS cluster on Outposts. This argument is a list of arns, but only a single Outpost ARN is supported currently.

### storage_config

The `storage_config` configuration block supports the following arguments:

* `block_storage` - (Optional) Configuration block with block storage configuration for the cluster. [Detailed](#block_storage) below.

### block_storage

The `block_storage` configuration block supports the following arguments:

* `enabled` - (Optional) Indicates if the block storage capability is enabled on your EKS Auto Mode cluster. If the block storage capability is enabled, EKS Auto Mode will create and delete block storage volumes in your Amazon Web Services account.

### upgrade_policy

The `upgrade_policy` configuration block supports the following arguments:

* `support_type` - (Optional) Support type to use for the cluster. If the cluster is set to `EXTENDED`, it will enter extended support at the end of standard support. If the cluster is set to `STANDARD`, it will be automatically upgraded at the end of standard support. Valid values are `EXTENDED`, `STANDARD`

### zonal_shift_config

The `zonal_shift_config` configuration block supports the following arguments:

* `enabled` - (Optional) Whether zonal shift is enabled for the cluster.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cluster.
* `certificate_authority` - Attribute block containing `certificate-authority-data` for your cluster. Detailed below.
* `cluster_id` - The ID of your local Amazon EKS cluster on the AWS Outpost. This attribute isn't available for an AWS EKS cluster on AWS cloud.
* `created_at` - Unix epoch timestamp in seconds for when the cluster was created.
* `endpoint` - Endpoint for your Kubernetes API server.
* `id` - Name of the cluster.
* `identity` - Attribute block containing identity provider information for your cluster. Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019. Detailed below.
* `platform_version` - Platform version for the cluster.
* `status` - Status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### certificate_authority

* `data` - Base64 encoded certificate data required to communicate with your cluster. Add this to the `certificate-authority-data` section of the `kubeconfig` file for your cluster.

### identity

* `oidc` - Nested block containing [OpenID Connect](https://openid.net/connect/) identity provider information for the cluster. Detailed below.

### oidc

* `issuer` - Issuer URL for the OpenID Connect identity provider.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `60m`)
Note that the `update` timeout is used separately for both `version` and `vpc_config` update timeouts.
* `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EKS Clusters using the `name`. For example:

```terraform
import {
  to = aws_eks_cluster.my_cluster
  id = "my_cluster"
}
```

Using `terraform import`, import EKS Clusters using the `name`. For example:

```console
% terraform import aws_eks_cluster.my_cluster my_cluster
```
