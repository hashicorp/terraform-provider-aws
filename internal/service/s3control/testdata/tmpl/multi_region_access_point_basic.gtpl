resource "aws_s3control_multi_region_access_point" "test" {
{{- template "region" }}
  details {
    name = var.rName

    region {
      bucket = aws_s3_bucket.test.id
    }
  }
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}
