resource "aws_sagemaker_user_profile" "test" {
{{- template "region" }}
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = var.rName

{{- template "tags" . }}
}

# testAccUserProfileConfig_base

resource "aws_sagemaker_domain" "test" {
{{- template "region" }}
  domain_name = var.rName
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = aws_subnet.test[*].id

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  retention_policy {
    home_efs_file_system = "Delete"
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
