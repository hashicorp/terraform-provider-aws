resource "aws_m2_application" "test" {
  name        = var.rName
  engine_type = "bluage"
  definition {
    content = templatefile("test-fixtures/application-definition.json", { s3_bucket = aws_s3_bucket.test.id, version = "v1" })
  }
{{- template "tags" . }}

  depends_on = [aws_s3_object.test]
}

resource "aws_s3_bucket" "test" {
  bucket = var.rName
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "v1/PlanetsDemo-v1.zip"
  source = "test-fixtures/PlanetsDemo-v1.zip"
}
