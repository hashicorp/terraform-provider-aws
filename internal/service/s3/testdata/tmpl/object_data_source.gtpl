data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
