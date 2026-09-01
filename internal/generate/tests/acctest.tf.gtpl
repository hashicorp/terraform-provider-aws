{{ define "acctest.ConfigVPCWithSubnets" -}}
# acctest.ConfigVPCWithSubnets(rName, {{ . }})

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}

{{ template "acctest.ConfigSubnets" . }}
{{- end }}

{{ define "acctest.ConfigSubnets" -}}
# acctest.ConfigSubnets(rName, {{ . }})

resource "aws_subnet" "test" {
{{- template "region" }}
  count = {{ . }}

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
{{- end }}

{{ define "acctest.ConfigVPCWithSubnetsIPv6" -}}
# acctest.ConfigVPCWithSubnetsIPv6(rName, {{ . }})

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true
}

{{ template "acctest.ConfigSubnetsIPv6" . }}
{{- end }}

{{ define "acctest.ConfigSubnetsIPv6" -}}
# acctest.ConfigSubnetsIPv6(rName, {{ . }})

resource "aws_subnet" "test" {
{{- template "region" }}
  count = {{ . }}

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]

  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
{{- end }}

{{ define "acctest.ConfigAvailableAZsNoOptIn" -}}
# acctest.ConfigAvailableAZsNoOptIn

data "aws_availability_zones" "available" {
{{- template "region" }}
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
{{- end }}

{{ define "acctest.ConfigAvailableAZsNoOptInExclude" -}}
# acctest.ConfigAvailableAZsNoOptInExclude

data "aws_availability_zones" "available" {
{{- template "region" }}
  exclude_zone_ids = local.exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
{{- end }}

{{ define "acctest.ConfigAvailableAZsNoOptInDefaultExclude" -}}
# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
{{- template "region" }}
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}
{{- end }}

{{ define "acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI" -}}
# acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI

{{ template "acctest.configLatestAmazonLinux2HVMEBSAMI" "x86_64" }}
{{- end }}

{{ define "acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI" -}}
# acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI

{{ template "acctest.configLatestAmazonLinux2HVMEBSAMI" "arm64" }}
{{- end }}

{{ define "acctest.configLatestAmazonLinux2HVMEBSAMI" -}}
# acctest.configLatestAmazonLinux2HVMEBSAMI("{{ . }}")

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-{{ . }}" {
{{- template "region" }}
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
    values = ["{{ . }}"]
  }
}
{{- end }}

{{ define "acctest.ConfigAlternateAccountProvider" -}}
provider "awsalternate" {
  access_key = var.AWS_ALTERNATE_ACCESS_KEY_ID
  profile    = var.AWS_ALTERNATE_PROFILE
  secret_key = var.AWS_ALTERNATE_SECRET_ACCESS_KEY
}

variable "AWS_ALTERNATE_ACCESS_KEY_ID" {
  type     = string
  nullable = true
  default  = null
}

variable "AWS_ALTERNATE_PROFILE" {
  type     = string
  nullable = true
  default  = null
}

variable "AWS_ALTERNATE_SECRET_ACCESS_KEY" {
  type     = string
  nullable = true
  default  = null
}
{{- end }}

{{ define "acctest.ConfigLambdaBase" -}}
{{/* Does not include the resource "aws_subnet.subnet_for_lambda_az2" */}}
# acctest.ConfigLambdaBase

resource "aws_iam_role" "iam_for_lambda" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = var.rName
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface",
        "ec2:AssignPrivateIpAddresses",
        "ec2:UnassignPrivateIpAddresses"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "SNS:Publish"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_vpc" "vpc_for_lambda" {
{{- template "region" }}
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true
}

resource "aws_subnet" "subnet_for_lambda" {
{{- template "region" }}
  vpc_id                          = aws_vpc.vpc_for_lambda.id
  cidr_block                      = cidrsubnet(aws_vpc.vpc_for_lambda.cidr_block, 8, 1)
  availability_zone               = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.vpc_for_lambda.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true
}

resource "aws_security_group" "sg_for_lambda" {
{{- template "region" }}
  name        = var.rName
  description = "Allow all inbound traffic for lambda test"
  vpc_id      = aws_vpc.vpc_for_lambda.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

data "aws_partition" "current" {}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
{{- end }}
