# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_eks_access_entry" "test" {
  provider = aws

  config {
    cluster_name          = aws_eks_cluster.test.name
    associated_policy_arn = one([for ap in data.aws_eks_access_policies.test.access_policies : ap.arn if ap.name == local.access_policy_names[0]])
  }
}
