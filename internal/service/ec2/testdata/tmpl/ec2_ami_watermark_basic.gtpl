resource "aws_ami_watermark" "test" {
{{- template "region" }}
  image_id       = aws_ami_copy.test.id
  watermark_name = var.rName
}

resource "aws_ami_copy" "test" {
{{- template "region" }}
  description       = var.rName
  name              = var.rName
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.region
}

{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" }}
