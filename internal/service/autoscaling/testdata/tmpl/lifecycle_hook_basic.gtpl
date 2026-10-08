resource "aws_autoscaling_lifecycle_hook" "test" {
{{- template "region" }}
  name                   = "${var.rName}-hook"
  autoscaling_group_name = aws_autoscaling_group.test.name
  default_result         = "CONTINUE"
  heartbeat_timeout      = 2000
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_LAUNCHING"

  notification_metadata = <<EOF
{
  "Key": "Value"
}
EOF

  notification_target_arn = aws_sqs_queue.test.arn
  role_arn                = aws_iam_role.test.arn
}

resource "aws_autoscaling_group" "test" {
{{- template "region" }}
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "${var.rName}-policy"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.test.name
}

resource "aws_launch_configuration" "test" {
{{- template "region" }}
  name          = var.rName
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}

resource "aws_sqs_queue" "test" {
{{- template "region" }}
  name                      = var.rName
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version" : "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": ["sts:AssumeRole"]
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version" : "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["sqs:SendMessage", "sqs:GetQueueUrl", "sns:Publish"],
    "Resource": ["${aws_sqs_queue.test.arn}"]
  }]
}
EOF
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
{{ template "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" }}
