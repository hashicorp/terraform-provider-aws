# Authenticated relays require a MailManager ingress point with SMTP credentials.
#
# The configuration below is a port of this guide to Terraform.
# Ref: https://docs.aws.amazon.com/ses/latest/dg/eb-relay.html#eb-relay-create-console.

resource "aws_mailmanager_relay" "test" {
  name        = var.rName
  server_name = aws_mailmanager_ingress_point.test.a_record
  server_port = 587

  authentication {
    secret_arn = aws_secretsmanager_secret_version.test.secret_arn
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_mailmanager_traffic_policy" "test" {
  default_action = "ALLOW"
  name           = var.rName

  policy_statement {
    action = "DENY"

    condition {
      ip_expression {
        operator = "CIDR_MATCHES"
        values   = ["192.0.2.0/24"]

        evaluate {
          attribute = "SENDER_IP"
        }
      }
    }
  }
}

resource "aws_mailmanager_rule_set" "test" {
  name = var.rName

  rule {
    action {
      add_header {
        header_name  = "X-Test"
        header_value = "example"
      }
    }
  }
}

resource "aws_mailmanager_ingress_point" "test" {
  name              = var.rName
  type              = "AUTH"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id

  ingress_point_configuration {
    smtp_password_wo         = var.password
    smtp_password_wo_version = 0
  }
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ses.amazonaws.com"
      },
      "Action": [
        "kms:Decrypt",
        "kms:DescribeKey"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "kms:ViaService": "secretsmanager.${data.aws_region.current.region}.amazonaws.com",
          "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
        },
        "ArnLike": {
					"aws:SourceArn": "arn:${data.aws_partition.current.partition}:ses:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:mailmanager-smtp-relay/*"
        }
      }
    }
  ]
}
EOF
}

resource "aws_secretsmanager_secret" "test" {
  name       = var.rName
  kms_key_id = aws_kms_key.test.arn

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service": "ses.amazonaws.com"
			},
			"Action": [
				"secretsmanager:GetSecretValue",
				"secretsmanager:DescribeSecret"
			],
			"Resource": "*",
			"Condition": {
				"StringEquals": {
					"aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
				},
				"ArnLike": {
					"aws:SourceArn": "arn:${data.aws_partition.current.partition}:ses:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:mailmanager-smtp-relay/*"
				}
			}
		}
	]
}
EOF
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = aws_mailmanager_ingress_point.test.id, password = var.password })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "password" {
  description = "SMTP password"
  type        = string
  sensitive   = true
  nullable    = false
}
