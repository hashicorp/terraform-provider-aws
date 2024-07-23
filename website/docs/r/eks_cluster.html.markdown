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

### Basic Usage

```terraform
resource "aws_eks_cluster" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn

  vpc_config {
    subnet_ids = [aws_subnet.example1.id, aws_subnet.example2.id]
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.example-AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.example-AmazonEKSVPCResourceController,
  ]
}

output "endpoint" {
  value = aws_eks_cluster.example.endpoint
}

output "kubeconfig-certificate-authority-data" {
  value = aws_eks_cluster.example.certificate_authority[0].data
}
```

### Example IAM Role for EKS Cluster

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["eks.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "eks-cluster-example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "example-AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.example.name
}

# Optionally, enable Security Groups for Pods
# Reference: https://docs.aws.amazon.com/eks/latest/userguide/security-groups-for-pods.html
resource "aws_iam_role_policy_attachment" "example-AmazonEKSVPCResourceController" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.example.name
}
```

### Enabling Control Plane Logging

[EKS Control Plane Logging](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html) can be enabled via the `enabled_cluster_log_types` argument. To manage the CloudWatch Log Group retention period, the [`aws_cloudwatch_log_group` resource](/docs/providers/aws/r/cloudwatch_log_group.html) can be used.

-> The below configuration uses [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html) to prevent ordering issues with EKS automatically creating the log group first and a variable for naming consistency. Other ordering and naming methodologies may be more appropriate for your environment.

```terraform
variable "cluster_name" {
  default = "example"
  type    = string
}

resource "aws_eks_cluster" "example" {
  depends_on = [aws_cloudwatch_log_group.example]

  enabled_cluster_log_types = ["api", "audit"]
  name                      = var.cluster_name

  # ... other configuration ...
}

resource "aws_cloudwatch_log_group" "example" {
  # The log group name format is /aws/eks/<cluster-name>/cluster
  # Reference: https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html
  name              = "/aws/eks/${var.cluster_name}/cluster"
  retention_in_days = 7

  # ... potentially other configuration ...
}
```

### Enabling IAM Roles for Service Accounts

For more information about this feature, see the [EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html).

```terraform
resource "aws_eks_cluster" "example" {
  # ... other configuration ...
}

data "tls_certificate" "example" {
  url = aws_eks_cluster.example.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "example" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.example.certificates[0].sha1_fingerprint]
  url             = data.tls_certificate.example.url
}

data "aws_iam_policy_document" "example_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "${replace(aws_iam_openid_connect_provider.example.url, "https://", "")}:sub"
      values   = ["system:serviceaccount:kube-system:aws-node"]
    }

    principals {
      identifiers = [aws_iam_openid_connect_provider.example.arn]
      type        = "Federated"
    }
  }
}

resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_assume_role_policy.json
  name               = "example"
}
```

### EKS Cluster on AWS Outpost

[Creating a local Amazon EKS cluster on an AWS Outpost](https://docs.aws.amazon.com/eks/latest/userguide/create-cluster-outpost.html)

```terraform
resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_assume_role_policy.json
  name               = "example"
}

resource "aws_eks_cluster" "example" {
  name     = "example-cluster"
  role_arn = aws_iam_role.example.arn

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = false
    # ... other configuration ...
  }

  outpost_config {
    control_plane_instance_type = "m5d.large"
    outpost_arns                = [data.aws_outposts_outpost.example.arn]
  }
}
```

### EKS Cluster with Access Config

```terraform
resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_assume_role_policy.json
  name               = "example"
}

resource "aws_eks_cluster" "example" {
  name     = "example-cluster"
  role_arn = aws_iam_role.example.arn

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = false
    # ... other configuration ...
  }

  access_config {
    authentication_mode                         = "CONFIG_MAP"
    bootstrap_cluster_creator_admin_permissions = true
  }
}
```

After adding inline IAM Policies (e.g., [`aws_iam_role_policy` resource](/docs/providers/aws/r/iam_role_policy.html)) or attaching IAM Policies (e.g., [`aws_iam_policy` resource](/docs/providers/aws/r/iam_policy.html) and [`aws_iam_role_policy_attachment` resource](/docs/providers/aws/r/iam_role_policy_attachment.html)) with the desired permissions to the IAM Role, annotate the Kubernetes service account (e.g., [`kubernetes_service_account` resource](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_account)) and recreate any pods.

## Argument Reference

The following arguments are required:

* `name` – (Required) Name of the cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\-_]*$`).
* `role_arn` - (Required) ARN of the IAM role that provides permissions for the Kubernetes control plane to make calls to AWS API operations on your behalf. Ensure the resource configuration includes explicit dependencies on the IAM Role permissions by adding [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html) if using the [`aws_iam_role_policy` resource](/docs/providers/aws/r/iam_role_policy.html) or [`aws_iam_role_policy_attachment` resource](/docs/providers/aws/r/iam_role_policy_attachment.html), otherwise EKS cannot delete EKS managed EC2 infrastructure such as Security Groups on EKS Cluster deletion.
* `vpc_config` - (Required) Configuration block for the VPC associated with your cluster. Amazon EKS VPC resources have specific requirements to work properly with Kubernetes. For more information, see [Cluster VPC Considerations](https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html) and [Cluster Security Group Considerations](https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html) in the Amazon EKS User Guide. Detailed below. Also contains attributes detailed in the Attributes section.

The following arguments are optional:

* `access_config` - (Optional) Configuration block for the access config associated with your cluster, see [Amazon EKS Access Entries](https://docs.aws.amazon.com/eks/latest/userguide/access-entries.html).
* `bootstrap_self_managed_addons` - (Optional) Install default unmanaged add-ons, such as `aws-cni`, `kube-proxy`, and CoreDNS during cluster creation. If `false`, you must manually install desired add-ons. Changing this value will force a new cluster to be created. Defaults to `true`.
* `enabled_cluster_log_types` - (Optional) List of the desired control plane logging to enable. For more information, see [Amazon EKS Control Plane Logging](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html).
* `encryption_config` - (Optional) Configuration block with encryption configuration for the cluster. Only available on Kubernetes 1.13 and above clusters created after March 6, 2020. Detailed below.
* `kubernetes_network_config` - (Optional) Configuration block with kubernetes network configuration for the cluster. Detailed below. If removed, Terraform will only perform drift detection if a configuration value is provided.
* `outpost_config` - (Optional) Configuration block representing the configuration of your local Amazon EKS cluster on an AWS Outpost. This block isn't available for creating Amazon EKS clusters on the AWS cloud.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `version` – (Optional) Desired Kubernetes master version. If you do not specify a value, the latest available version at resource creation is used and no upgrades will occur except those automatically triggered by EKS. The value must be configured and increased to upgrade the version when desired. Downgrades are not supported by EKS.

### access_config

The `access_config` configuration block supports the following arguments:

* `authentication_mode` - (Optional) The authentication mode for the cluster. Valid values are `CONFIG_MAP`, `API` or `API_AND_CONFIG_MAP`
* `bootstrap_cluster_creator_admin_permissions` - (Optional) Whether or not to bootstrap the access config values to the cluster. Default is `false`.

### encryption_config

The `encryption_config` configuration block supports the following arguments:

* `provider` - (Required) Configuration block with provider for encryption. Detailed below.
* `resources` - (Required) List of strings with resources to be encrypted. Valid values: `secrets`.

#### provider

The `provider` configuration block supports the following arguments:

* `key_arn` - (Required) ARN of the Key Management Service (KMS) customer master key (CMK). The CMK must be symmetric, created in the same region as the cluster, and if the CMK was created in a different account, the user must have access to the CMK. For more information, see [Allowing Users in Other Accounts to Use a CMK in the AWS Key Management Service Developer Guide](https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-modifying-external-accounts.html).

### vpc_config Arguments

* `endpoint_private_access` - (Optional) Whether the Amazon EKS private API server endpoint is enabled. Default is `false`.
* `endpoint_public_access` - (Optional) Whether the Amazon EKS public API server endpoint is enabled. Default is `true`.
* `public_access_cidrs` - (Optional) List of CIDR blocks. Indicates which CIDR blocks can access the Amazon EKS public API server endpoint when enabled. EKS defaults this to a list with `0.0.0.0/0`. Terraform will only perform drift detection of its value when present in a configuration.
* `security_group_ids` – (Optional) List of security group IDs for the cross-account elastic network interfaces that Amazon EKS creates to use to allow communication between your worker nodes and the Kubernetes control plane.
* `subnet_ids` – (Required) List of subnet IDs. Must be in at least two different availability zones. Amazon EKS creates cross-account elastic network interfaces in these subnets to allow communication between your worker nodes and the Kubernetes control plane.

### kubernetes_network_config

The `kubernetes_network_config` configuration block supports the following arguments:

* `service_ipv4_cidr` - (Optional) The CIDR block to assign Kubernetes pod and service IP addresses from. If you don't specify a block, Kubernetes assigns addresses from either the 10.100.0.0/16 or 172.20.0.0/16 CIDR blocks. We recommend that you specify a block that does not overlap with resources in other networks that are peered or connected to your VPC. You can only specify a custom CIDR block when you create a cluster, changing this value will force a new cluster to be created. The block must meet the following requirements:

    * Within one of the following private IP address blocks: 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16.

    * Doesn't overlap with any CIDR block assigned to the VPC that you selected for VPC.

    * Between /24 and /12.
* `ip_family` - (Optional) The IP family used to assign Kubernetes pod and service addresses. Valid values are `ipv4` (default) and `ipv6`. You can only specify an IP family when you create a cluster, changing this value will force a new cluster to be created.

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

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cluster.
* `certificate_authority` - Attribute block containing `certificate-authority-data` for your cluster. Detailed below.
* `cluster_id` - The ID of your local Amazon EKS cluster on the AWS Outpost. This attribute isn't available for an AWS EKS cluster on AWS cloud.
* `created_at` - Unix epoch timestamp in seconds for when the cluster was created.
* `endpoint` - Endpoint for your Kubernetes API server.
* `id` - Name of the cluster.
* `identity` - Attribute block containing identity provider information for your cluster. Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019. Detailed below.
* `kubernetes_network_config.service_ipv6_cidr` - The CIDR block that Kubernetes pod and service IP addresses are assigned from if you specified `ipv6` for ipFamily when you created the cluster. Kubernetes assigns service addresses from the unique local address range (fc00::/7) because you can't specify a custom IPv6 CIDR block when you create the cluster.
* `platform_version` - Platform version for the cluster.
* `status` - Status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_config` - Configuration block _argument_ that also includes attributes for the VPC associated with your cluster. Detailed below.

### certificate_authority

* `data` - Base64 encoded certificate data required to communicate with your cluster. Add this to the `certificate-authority-data` section of the `kubeconfig` file for your cluster.

### identity

* `oidc` - Nested block containing [OpenID Connect](https://openid.net/connect/) identity provider information for the cluster. Detailed below.

### oidc

* `issuer` - Issuer URL for the OpenID Connect identity provider.

### vpc_config Attributes

* `cluster_security_group_id` - Cluster security group that was created by Amazon EKS for the cluster. Managed node groups use this security group for control-plane-to-data-plane communication.
* `vpc_id` - ID of the VPC associated with your cluster.

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
