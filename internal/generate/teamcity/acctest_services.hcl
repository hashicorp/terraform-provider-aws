# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

service "appautoscaling" {
  vpc_lock = true
}

service "appstream" {
  vpc_lock    = true
  parallelism = 10
}

service "autoscaling" {
  vpc_lock = true
}

service "batch" {
  vpc_lock = true
}

service "cloudformation" {
  vpc_lock = true
}

service "cloudhsmv2" {
  vpc_lock = true
}

service "comprehend" {
  parallelism = 10
}

service "cur" {
  region = "us-east-1"
}

service "datasync" {
  vpc_lock = true
}

service "directconnect" {
  vpc_lock = true
}

service "dms" {
  vpc_lock = true
}

service "docdb" {
  vpc_lock = true
}

service "ds" {
  vpc_lock = true
}

service "ec2" {
  vpc_lock = true
}

service "ecrpublic" {
  region = "us-east-1"
}

service "efs" {
  vpc_lock = true
}

service "eks" {
  vpc_lock = true
}

service "elasticache" {
  vpc_lock = true
}

service "elasticbeanstalk" {
  vpc_lock = true
}

service "elasticsearch" {
  vpc_lock = true
}

service "elb" {
  vpc_lock = true
}

service "elbv2" {
  vpc_lock = true
}

service "emr" {
  vpc_lock = true
}

service "fms" {
  region = "us-east-1"
}

service "fsx" {
  vpc_lock = true
}

service "kafka" {
  vpc_lock = true
}

service "kendra" {
  skip = true
}

service "kinesisanalytics" {
  skip = true
}

service "kinesisanalyticsv2" {
  skip = true
}

service "lambda" {
  vpc_lock = true
}

service "lightsail" {
  region = "us-east-1"
}

service "mq" {
  vpc_lock = true
}

service "mwaa" {
  vpc_lock = true
}

service "networkfirewall" {
  vpc_lock = true
}

service "opsworks" {
  vpc_lock = true
}

service "pricing" {
  region = "us-east-1"
}

service "rds" {
  vpc_lock = true
}

service "redshift" {
  vpc_lock = true
}

service "route53" {
  vpc_lock = true
}

service "route53resolver" {
  vpc_lock = true
}

service "sagemaker" {
  vpc_lock = true
}

service "servicediscovery" {
  vpc_lock = true
}

service "ssm" {
  vpc_lock = true
}

service "storagegateway" {
  vpc_lock = true
}

service "synthetics" {
  parallelism = 10
}

service "transfer" {
  vpc_lock = true
}

service "waf" {
  region = "us-east-1"
}

service "workspaces" {
  # Needed for logging configuration tests
  vpc_lock = true
}
