# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

service "amp" {
  # The maximum scrapers per region quota is fixed at 10
  parallelism = 10
}

service "appautoscaling" {
  vpc_lock = true
}

service "apigateway" {
  vpc_lock = true
}

service "apigatewayv2" {
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

service "deploy" {
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
  vpc_lock         = true
  pattern_override = "TestAccEC2"
  exclude_pattern  = "TestAccEC2EBS|TestAccEC2Outposts"
}

service "ec2ebs" {
  vpc_lock                   = true
  pattern_override           = "TestAccEC2EBS"
  split_package_real_package = "ec2"
}

service "ec2outposts" {
  vpc_lock                   = true
  pattern_override           = "TestAccEC2Outposts"
  split_package_real_package = "ec2"
}

service "ecrpublic" {
  region = "us-east-1"
}

service "ecs" {
  vpc_lock = true
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

service "imagebuilder" {
  vpc_lock = true
}

service "ipam" {
  vpc_lock                   = true
  pattern_override           = "TestAccIPAM"
  split_package_real_package = "ec2"
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

service "networkmanager" {
  vpc_lock = true
}

service "opensearch" {
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

service "transitgateway" {
  vpc_lock                   = true
  pattern_override           = "TestAccTransitGateway"
  split_package_real_package = "ec2"
}

service "verifiedaccess" {
  vpc_lock                   = true
  pattern_override           = "TestAccVerifiedAccess"
  split_package_real_package = "ec2"
}

service "vpc" {
  vpc_lock                   = true
  pattern_override           = "TestAccVPC"
  split_package_real_package = "ec2"
}

service "vpnclient" {
  vpc_lock                   = true
  pattern_override           = "TestAccClientVPN"
  split_package_real_package = "ec2"
}

service "vpnsite" {
  vpc_lock                   = true
  pattern_override           = "TestAccSiteVPN"
  split_package_real_package = "ec2"
}

service "waf" {
  region = "us-east-1"
}

service "wavelength" {
  vpc_lock                   = true
  pattern_override           = "TestAccWavelength"
  split_package_real_package = "ec2"
}

service "workspaces" {
  # Needed for logging configuration tests
  vpc_lock = true
}
