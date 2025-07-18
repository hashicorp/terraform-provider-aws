resource "aws_glue_registry" "test" {
  provider = aws

  registry_name = var.rName
}
