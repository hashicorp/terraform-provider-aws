resource "aws_ami_launch_permission" "test" {
{{- template "region" }}
  account_id = data.aws_caller_identity.current.account_id
  image_id   = aws_ami_copy.test.id
}

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-x86_64" {
{{- template "region" }}
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*-x86_64-ebs"]
  }
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {
{{- template "region" -}}
}

resource "aws_ami_copy" "test" {
{{- template "region" }}
  description       = var.rName
  name              = var.rName
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_region.current.region
}
