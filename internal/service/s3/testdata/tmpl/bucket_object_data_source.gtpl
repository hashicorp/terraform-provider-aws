data "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket_object.test.bucket
  key    = aws_s3_bucket_object.test.key
}
