resource "aws_autoscaling_group" "test" {
{{- template "region" }}
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = var.rName
  launch_configuration = aws_launch_configuration.test.name
}

resource "aws_launch_configuration" "test" {
{{- template "region" }}
  name_prefix   = var.rName
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  lifecycle {
    create_before_destroy = true
  }
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" }}
