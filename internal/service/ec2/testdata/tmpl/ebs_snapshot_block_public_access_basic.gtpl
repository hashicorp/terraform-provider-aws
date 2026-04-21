resource "aws_ebs_snapshot_block_public_access" "test" {
{{- template "region" }}
  state = "block-all-sharing"
}
