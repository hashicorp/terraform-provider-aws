provider "aws" {
  region  = "us-west-2"
  profile = "provider5"
}

resource "aws_batch_job_queue" "test" {
  count = 3

  name     = "list-test-${count.index}"
  priority = 1
  state    = "DISABLED"

  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }
}

resource "aws_batch_compute_environment" "test" {
  name         = "list-test"
  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"

  depends_on = [aws_iam_role_policy_attachment.batch_service]
}

data "aws_partition" "current" {}

resource "aws_iam_role" "batch_service" {
  name_prefix = "tf-test-list-batch-service"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "batch.${data.aws_partition.current.dns_suffix}"
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "batch_service" {
  role       = aws_iam_role.batch_service.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ecs_instance" {
  name_prefix = "tf-test-list-ecs-instance"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
        }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance" {
  role       = aws_iam_role.ecs_instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance" {
  name = aws_iam_role.ecs_instance.name
  role = aws_iam_role_policy_attachment.ecs_instance.role
}

resource "aws_instance" "no_name" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.nano"
}

resource "aws_instance" "named" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.nano"

  tags = {
    Name = "list-test"
  }
}

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-arm64" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "architecture"
    values = ["arm64"]
  }
}
