resource "aws_sagemaker_endpoint_configuration" "test" {
{{- template "region" }}
  name = var.rName

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

{{- template "tags" . }}
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
{{- template "region" }}
  repository_name = "kmeans"
}

resource "aws_sagemaker_model" "test" {
{{- template "region" }}
  name               = var.rName
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
