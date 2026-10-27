# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lambdaweb_function_revision" "test" {
  depends_on = [aws_s3_bucket_policy.test, aws_s3_bucket_versioning.test, aws_iam_role_policy_attachment.test]

  function_name = aws_lambdaweb_function.test.function_name

  build_config {
    runtime_config {
      runtime = "nodejs24.x"
    }

    code_config {
      s3_object {
        bucket = aws_s3_object.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  service_config {
    egress_network_connector_arn = aws_lambdacore_network_connector.test.latest_version_arn
    execution_role_arn           = aws_iam_role.test.arn
  }
}

resource "aws_lambdacore_network_connector" "test" {
  name          = var.rName
  operator_role = aws_iam_role.operator.arn

  configuration {
    vpc_egress_configuration {
      associated_compute_resource_types = ["WebFunction"]
      network_protocol                  = "IPv4"
      subnet_ids                        = aws_subnet.test[*].id
      security_group_ids                = [aws_security_group.test.id]
    }
  }

  depends_on = [aws_iam_role_policy.operator]
}

# acctest.ConfigVPCWithSubnets(rName, 2)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

# acctest.ConfigSubnets(rName, 2)

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
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

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_iam_role" "operator" {
  name = "${var.rName}-operator"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "network-connectors.lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "operator" {
  name = "${var.rName}-operator"
  role = aws_iam_role.operator.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "CreateENI"
        Effect = "Allow"
        Action = "ec2:CreateNetworkInterface"
        Resource = [
          "arn:${data.aws_partition.current.partition}:ec2:*:*:network-interface/*",
          "arn:${data.aws_partition.current.partition}:ec2:*:*:subnet/*",
          "arn:${data.aws_partition.current.partition}:ec2:*:*:security-group/*",
        ]
      },
      {
        Sid      = "TagENI"
        Effect   = "Allow"
        Action   = "ec2:CreateTags"
        Resource = "arn:${data.aws_partition.current.partition}:ec2:*:*:network-interface/*"
        Condition = {
          StringEquals = {
            "ec2:ManagedResourceOperator" = "network-connectors.lambda.amazonaws.com"
          }
        }
      },
    ]
  })
}

resource "aws_lambdaweb_function" "test" {
  function_name = var.rName
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowLambdaAccess"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      Action   = ["s3:GetObject", "s3:GetObjectVersion"]
      Resource = "${aws_s3_bucket.test.arn}/*"
    }]
  })
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket_versioning.test.bucket
  key    = "function.zip"
  source = "test-fixtures/function.zip"
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
