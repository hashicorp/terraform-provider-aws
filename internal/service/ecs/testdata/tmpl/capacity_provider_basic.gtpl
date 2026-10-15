resource "aws_ecs_capacity_provider" "test" {
{{- template "region" }}
  name = var.rName

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
{{- template "tags" . }}
}

# testAccCapacityProviderConfig_base

resource "aws_launch_template" "test" {
{{- template "region" }}
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  name          = var.rName
}

resource "aws_autoscaling_group" "test" {
{{- template "region" }}
  availability_zones = data.aws_availability_zones.available.names
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = var.rName

  launch_template {
    id = aws_launch_template.test.id
  }

  tag {
    key                 = "Name"
    value               = var.rName
    propagate_at_launch = true
  }

  lifecycle {
    ignore_changes = [
      tag,
    ]
  }
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}

{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" }}
