data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["iot.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSIoTThingsRegistration"
}

data "aws_iam_policy_document" "device" {
  statement {
    actions   = ["iot:Subscribe"]
    resources = ["*"]
  }
}

resource "aws_iot_policy" "test" {
{{- template "region" }}
  name   = var.rName
  policy = data.aws_iam_policy_document.device.json
}

resource "aws_iot_provisioning_template" "test" {
{{- template "region" }}
  name                  = var.rName
  provisioning_role_arn = aws_iam_role.test.arn

  template_body = jsonencode({
    Parameters = {
      SerialNumber = { Type = "String" }
    }

    Resources = {
      certificate = {
        Properties = {
          CertificateId = { Ref = "AWS::IoT::Certificate::Id" }
          Status        = "Active"
        }
        Type = "AWS::IoT::Certificate"
      }

      policy = {
        Properties = {
          PolicyName = aws_iot_policy.test.name
        }
        Type = "AWS::IoT::Policy"
      }
    }
  })

{{- template "tags" . }}
}
