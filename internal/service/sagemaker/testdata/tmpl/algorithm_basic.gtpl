resource "aws_sagemaker_algorithm" "test" {
{{- template "region" }}
  algorithm_name = var.rName

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }

{{- template "tags" . }}
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
{{- template "region" }}
  repository_name = "linear-learner"
  image_tag       = "1"
}