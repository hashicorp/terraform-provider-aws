resource "aws_ec2_image_block_public_access" "test" {
  state = "block-new-sharing"
}
