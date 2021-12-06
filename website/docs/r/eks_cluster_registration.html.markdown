---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_cluster_registration"
description: |-
  Registers an external Kubernetes Cluster with EKS
---

# Resource: aws_eks_cluster_registration

Registers an external Kubernetes Cluster with EKS

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
resource "aws_iam_role" "example" {
  name = "eks-cluster-example"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
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

Only available on Kubernetes version 1.13 and 1.14 clusters created or upgraded on or after September 3, 2019. For more information about this feature, see the [EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html).

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
  url             = aws_eks_cluster.example.identity[0].oidc[0].issuer
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

After adding inline IAM Policies (e.g., [`aws_iam_role_policy` resource](/docs/providers/aws/r/iam_role_policy.html)) or attaching IAM Policies (e.g., [`aws_iam_policy` resource](/docs/providers/aws/r/iam_policy.html) and [`aws_iam_role_policy_attachment` resource](/docs/providers/aws/r/iam_role_policy_attachment.html)) with the desired permissions to the IAM Role, annotate the Kubernetes service account (e.g., [`kubernetes_service_account` resource](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_account)) and recreate any pods.

## Argument Reference

The following arguments are required:

* `name` – (Required) Name of the cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\-_]+$`).
* `connector_config` - (Required) The configuration settings required to connect the Kubernetes cluster to the Amazon EKS control plane.

The following arguments are optional:

* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### connector_config

The following arguments are supported in the `encryption_config` configuration block:

* `provider` - (Required) The cloud provider for the target cluster to connect.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the role that is authorized to request the connector configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the cluster.
* `connector_config` - (Required) The configuration settings required to connect the Kubernetes cluster to the Amazon EKS control plane.
* `created_at` - Unix epoch timestamp in seconds for when the cluster was created.
* `status` - Status of the EKS cluster. One of `ACTIVE`, `FAILED`, `PENDING`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

### connector_config

* `activation_id` - A unique code associated with the cluster for registration purposes.
* `activation_expiry` - The expiration time of the connected cluster. The cluster's YAML file must be applied through the native provider.

## Import

EKS Clusters can be imported using the `name`, e.g.,

```
$ terraform import aws_eks_cluster_registration.my_cluster my_cluster
```
