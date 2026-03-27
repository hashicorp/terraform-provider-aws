resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
{{- template "region" }}
  hyper_parameter_tuning_job_name = var.rName

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    resource_limits {
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      algorithm_name      = "xgboost"
      training_input_mode = "File"
    }

    input_data_config {
      channel_name = "train"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
        }
      }
    }

    output_data_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

{{- template "tags" . }}
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

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_s3_bucket" "test" {
  bucket        = "${var.rName}-hptj"
  force_destroy = true
}

