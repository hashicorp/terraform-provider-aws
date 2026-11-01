# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

output "hsm_ip_address" {
  value = aws_cloudhsm_v2_hsm.cloudhsm_v2_hsm.ip_address
}

output "cluster_data_certificate" {
  value = data.aws_cloudhsm_v2_cluster.cluster.cluster_certificates[0].cluster_csr
}
output "s3_bucket_arn" {
  value = aws_s3control_bucket.bucket_name.arn
}

output "s3_access_point_arn" {
  value = aws_s3_access_point.op_access_point.arn
}