resource "aws_autoscaling_policy" "test" {
{{- template "region" }}
  name                   = "${var.rName}-policy"
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 2
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_group" "test" {
{{- template "region" }}
  availability_zones   = slice(data.aws_availability_zones.available.names, 0, 2)
  name                 = "${var.rName}-group"
  max_size             = 0
  min_size             = 0
  force_delete         = true
  launch_configuration = aws_launch_configuration.test.name
}

resource "aws_launch_configuration" "test" {
{{- template "region" }}
  name          = var.rName
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" }}
