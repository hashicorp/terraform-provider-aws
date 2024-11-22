# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.12"
}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "AWSElasticDisasterRecoveryAgentRole" {
  name = "AWSElasticDisasterRecoveryAgentRole"
  path = "/service-role/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "drs.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetSourceIdentity"
        ]
        Condition = {
          StringLike = {
            "sts:SourceIdentity" = "s-*",
            "aws:SourceAccount"  = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_role" "AWSElasticDisasterRecoveryFailbackRole" {
  name = "AWSElasticDisasterRecoveryFailbackRole"
  path = "/service-role/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "drs.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetSourceIdentity"
        ]
        Condition = {
          StringLike = {
            "sts:SourceIdentity" = "i-*",
            "aws:SourceAccount"  = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_role" "AWSElasticDisasterRecoveryConversionServerRole" {
  name = "AWSElasticDisasterRecoveryConversionServerRole"
  path = "/service-role/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role" "AWSElasticDisasterRecoveryRecoveryInstanceRole" {
  name = "AWSElasticDisasterRecoveryRecoveryInstanceRole"
  path = "/service-role/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role" "AWSElasticDisasterRecoveryReplicationServerRole" {
  name = "AWSElasticDisasterRecoveryReplicationServerRole"
  path = "/service-role/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role" "AWSElasticDisasterRecoveryRecoveryInstanceWithLaunchActionsRole" {
  name = "AWSElasticDisasterRecoveryRecoveryInstanceWithLaunchActionsRole"
  path = "/service-role/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

data "aws_iam_policy" "AWSElasticDisasterRecoveryAgentPolicy" {
  name = "AWSElasticDisasterRecoveryAgentPolicy"
}

data "aws_iam_policy" "AWSElasticDisasterRecoveryFailbackPolicy" {
  name = "AWSElasticDisasterRecoveryFailbackPolicy"
}

data "aws_iam_policy" "AWSElasticDisasterRecoveryConversionServerPolicy" {
  name = "AWSElasticDisasterRecoveryConversionServerPolicy"
}

data "aws_iam_policy" "AWSElasticDisasterRecoveryRecoveryInstancePolicy" {
  name = "AWSElasticDisasterRecoveryRecoveryInstancePolicy"
}

data "aws_iam_policy" "AWSElasticDisasterRecoveryReplicationServerPolicy" {
  name = "AWSElasticDisasterRecoveryReplicationServerPolicy"
}

data "aws_iam_policy" "AmazonSSMManagedInstanceCore" {
  name = "AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role_policy_attachment" "AWSElasticDisasterRecoveryAgentRole" {
  role       = aws_iam_role.AWSElasticDisasterRecoveryAgentRole.name
  policy_arn = data.aws_iam_policy.AWSElasticDisasterRecoveryAgentPolicy.arn
}

resource "aws_iam_role_policy_attachment" "AWSElasticDisasterRecoveryFailbackRole" {
  role       = aws_iam_role.AWSElasticDisasterRecoveryFailbackRole.name
  policy_arn = data.aws_iam_policy.AWSElasticDisasterRecoveryFailbackPolicy.arn
}

resource "aws_iam_role_policy_attachment" "AWSElasticDisasterRecoveryConversionServerRole" {
  role       = aws_iam_role.AWSElasticDisasterRecoveryConversionServerRole.name
  policy_arn = data.aws_iam_policy.AWSElasticDisasterRecoveryConversionServerPolicy.arn
}

resource "aws_iam_role_policy_attachment" "AWSElasticDisasterRecoveryRecoveryInstanceRole" {
  role       = aws_iam_role.AWSElasticDisasterRecoveryRecoveryInstanceRole.name
  policy_arn = data.aws_iam_policy.AWSElasticDisasterRecoveryRecoveryInstancePolicy.arn
}

resource "aws_iam_role_policy_attachment" "AWSElasticDisasterRecoveryReplicationServerRole" {
  role       = aws_iam_role.AWSElasticDisasterRecoveryReplicationServerRole.name
  policy_arn = data.aws_iam_policy.AWSElasticDisasterRecoveryReplicationServerPolicy.arn
}

resource "aws_iam_role_policy_attachment" "AWSElasticDisasterRecoveryRecoveryInstanceWithLaunchActionsRole1" {
  role       = aws_iam_role.AWSElasticDisasterRecoveryRecoveryInstanceWithLaunchActionsRole.name
  policy_arn = data.aws_iam_policy.AWSElasticDisasterRecoveryRecoveryInstancePolicy.arn
}

resource "aws_iam_role_policy_attachment" "AWSElasticDisasterRecoveryRecoveryInstanceWithLaunchActionsRole2" {
  role       = aws_iam_role.AWSElasticDisasterRecoveryRecoveryInstanceWithLaunchActionsRole.name
  policy_arn = data.aws_iam_policy.AmazonSSMManagedInstanceCore.arn
}


resource "aws_iam_role" "AWSElasticDisasterRecoveryInitializerRole" {
  name = "AWSElasticDisasterRecoveryInitializerRole"
  path = "/"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.user_id
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_policy" "InitializePolicy" {
  name        = "InitializePolicy"
  description = "Policy for initializing the AWS Elastic Disaster Recovery service"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "iam:AttachRolePolicy"
        Resource = "*"
        Condition = {
          "ForAnyValue:ArnEquals" = {
            "iam:PolicyARN" = [
              data.aws_iam_policy.AWSElasticDisasterRecoveryAgentPolicy.arn,
              data.aws_iam_policy.AWSElasticDisasterRecoveryFailbackPolicy.arn,
              data.aws_iam_policy.AWSElasticDisasterRecoveryConversionServerPolicy.arn,
              data.aws_iam_policy.AWSElasticDisasterRecoveryRecoveryInstancePolicy.arn,
              data.aws_iam_policy.AWSElasticDisasterRecoveryReplicationServerPolicy.arn
            ]
          }
        }
      },
      {
        Effect   = "Allow"
        Action   = "iam:PassRole"
        Resource = "arn:aws:iam::*:role/*"
        Condition = {
          "ForAnyValue:StringLike" = {
            "iam:PassedToService" = [
              "ec2.amazonaws.com",
              "drs.amazonaws.com"
            ]
          }
        }
      },
      {
        Effect = "Allow"
        Action = [
          "drs:InitializeService",
          "drs:ListTagsForResource",
          "drs:GetReplicationConfiguration",
          "drs:CreateLaunchConfigurationTemplate",
          "drs:GetLaunchConfiguration",
          "drs:CreateReplicationConfigurationTemplate",
          "drs:*ReplicationConfigurationTemplate*",
          "iam:TagRole",
          "iam:CreateRole",
          "iam:GetServiceLinkedRoleDeletionStatus",
          "iam:ListAttachedRolePolicies",
          "iam:ListRolePolicies",
          "iam:GetRole",
          "iam:DeleteRole",
          "iam:DeleteServiceLinkedRole",
          "ec2:*",
          "sts:DecodeAuthorizationMessage",
        ]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = "iam:CreateServiceLinkedRole"
        Resource = "arn:aws:iam::*:role/aws-service-role/drs.amazonaws.com/AWSServiceRoleForElasticDisasterRecovery"
      },
      {
        Effect = "Allow"
        Action = [
          "iam:CreateInstanceProfile",
          "iam:ListInstanceProfilesForRole",
          "iam:GetInstanceProfile",
          "iam:ListInstanceProfiles",
          "iam:AddRoleToInstanceProfile"
        ]
        Resource = [
          "arn:aws:iam::*:instance-profile/*",
          "arn:aws:iam::*:role/*"
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "Initializer" {
  role       = aws_iam_role.AWSElasticDisasterRecoveryInitializerRole.name
  policy_arn = aws_iam_policy.InitializePolicy.arn
}
