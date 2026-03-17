resource "aws_ec2_image_block_public_access" "test" {
{{- template "region" }}
  state = "block-new-sharing"
}
