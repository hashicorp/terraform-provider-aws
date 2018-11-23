#
# Apply aws-auth config map to k8s cluster
#

resource "k8s_manifest" "awsauth" {
  content = "${local.config_map_aws_auth}"
}
