data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }
}

resource "aws_networkflowmonitor_monitor" "test" {
  monitor_name = var.rName
  scope_arn    = aws_networkflowmonitor_scope.test.arn

  local_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }

  remote_resources {
    type       = "AWS::EC2::VPC"
    identifier = aws_vpc.test.arn
  }
{{- template "tags" . }}
}