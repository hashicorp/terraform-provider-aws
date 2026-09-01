resource "aws_autoscaling_schedule" "test" {
{{- template "region" }}
  scheduled_action_name  = "${var.rName}-schedule"
  min_size               = 0
  max_size               = 1
  desired_capacity       = 0
  start_time             = var.startTime
  end_time               = var.endTime
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_group" "test" {
{{- template "region" }}
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "${var.rName}-group"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = var.rName
    propagate_at_launch = true
  }
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

variable "startTime" {
  description = "Schedule start time"
  type        = string
  nullable    = false
}

variable "endTime" {
  description = "Schedule end time"
  type        = string
  nullable    = false
}
