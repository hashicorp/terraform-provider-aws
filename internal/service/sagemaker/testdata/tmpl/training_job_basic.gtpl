resource "aws_sagemaker_training_job" "test" {
{{- template "region" }}
  training_job_name = var.rName
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/linear-learner:1"
  }

  output_data_config {
    kms_key_id     = aws_kms_key.test.arn
    s3_output_path = "s3://example-training-job-output/"
  }

  resource_config {
    instance_type  = "ml.m5.large"
    instance_count = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
{{- template "tags" . }}
}

data "aws_partition" "current" {}

data "aws_region" "current" {
{{- template "region" }}
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_kms_key" "test" {
{{- template "region" }}
  description = "KMS key for SageMaker training job"
}
