resource "aws_datazone_environment_blueprint" "test" {
{{- template "region" }}
  domain_identifier = aws_datazone_domain.test.id
  name              = var.rName
  description       = var.rName

  provisioning_properties {
    cloud_formation {
      template_url = "https://${aws_s3_object.template.bucket}.s3.amazonaws.com/${aws_s3_object.template.key}"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.domain_execution]
}

# The CloudFormation template must live in an S3 bucket in the domain account.
resource "aws_s3_bucket" "template" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "template" {
{{- template "region" }}
  bucket  = aws_s3_bucket.template.id
  key     = "template.yaml"
  content = <<-YAML
    Resources:
      Index:
        Type: AWS::OpenSearchServerless::Index
        Properties:
          IndexName: textIndex
          CollectionEndpoint: colTest
  YAML
}

# Custom CloudFormation blueprints require a V2 (SageMaker Unified Studio) domain.
resource "aws_datazone_domain" "test" {
{{- template "region" }}
  name                  = var.rName
  domain_version        = "V2"
  domain_execution_role = aws_iam_role.domain_execution.arn
  service_role          = aws_iam_role.domain_service.arn
  skip_deletion_check   = true
}

data "aws_caller_identity" "current" {}

# IAM role for Domain Execution
data "aws_iam_policy_document" "assume_role_domain_execution" {
  statement {
    actions = ["sts:AssumeRole", "sts:TagSession", "sts:SetContext"]

    principals {
      type        = "Service"
      identifiers = ["datazone.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
    condition {
      test     = "ForAllValues:StringLike"
      values   = ["datazone*"]
      variable = "aws:TagKeys"
    }
  }
}

resource "aws_iam_role" "domain_execution" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_domain_execution.json
  name               = "${var.rName}-exec"
}

data "aws_iam_policy" "domain_execution_role" {
  name = "SageMakerStudioDomainExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "domain_execution" {
  policy_arn = data.aws_iam_policy.domain_execution_role.arn
  role       = aws_iam_role.domain_execution.name
}

# IAM role for Domain Service
data "aws_iam_policy_document" "assume_role_domain_service" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["datazone.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
  }
}

resource "aws_iam_role" "domain_service" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_domain_service.json
  name               = "${var.rName}-svc"
}

data "aws_iam_policy" "domain_service_role" {
  name = "SageMakerStudioDomainServiceRolePolicy"
}

resource "aws_iam_role_policy_attachment" "domain_service" {
  policy_arn = data.aws_iam_policy.domain_service_role.arn
  role       = aws_iam_role.domain_service.name
}
