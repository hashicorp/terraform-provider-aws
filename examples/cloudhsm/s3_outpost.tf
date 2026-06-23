# s3_outpost.tf
resource "aws_s3control_bucket" "bucket_name" {
  bucket     = "test0001"
  outpost_id = var.outpost_id
}

resource "aws_s3_access_point" "op_access_point" {
  bucket = aws_s3control_bucket.bucket_name.id
  name   = "ap-test0001"

  vpc_configuration {
    vpc_id = var.vpc_id
  }
}

resource "aws_s3control_bucket_versioning" "backend_outpost_local" {
  bucket = aws_s3control_bucket.bucket_name.arn
  
  versioning_configuration {
    status = "Enabled"
  }
}