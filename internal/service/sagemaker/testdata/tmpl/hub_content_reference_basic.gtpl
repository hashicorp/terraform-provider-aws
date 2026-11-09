resource "aws_sagemaker_hub_content_reference" "test" {
{{- template "region" }}
  hub_name                         = aws_sagemaker_hub.test.hub_name
  hub_content_name                 = var.rName
  sagemaker_public_hub_content_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct"

{{- template "tags" . }}
}

data "aws_partition" "current" {}
data "aws_region" "current" {
{{- template "region" }}
}

resource "aws_sagemaker_hub" "test" {
{{- template "region" }}
  hub_name        = var.rName
  hub_description = var.rName
}
