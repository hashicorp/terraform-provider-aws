resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = var.rName
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  tagging_directive = "REPLACE"

{{- template "tags" . }}
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  key     = var.rName
  content = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

resource "aws_s3_bucket" "source" {
  bucket = "${var.rName}-source"

  force_destroy = true
}

resource "aws_s3_bucket" "target" {
  bucket = "${var.rName}-target"
}
