resource "aws_dataexchange_revision_assets" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  asset {
    import_assets_from_s3 {
      asset_source {
        bucket = aws_s3_object.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

{{- template "tags" . }}
}

resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = var.rName
  name        = var.rName
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "test"
  content = "test"
}
