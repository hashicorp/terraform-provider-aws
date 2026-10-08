resource "aws_launch_configuration" "test" {
{{- template "region" }}
  name          = var.rName
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}

{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" }}
