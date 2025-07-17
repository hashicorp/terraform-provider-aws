resource "aws_cloudfrontkeyvaluestore_key" "test" {
  key                 = var.rName
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn
  value               = var.rName
}

resource "aws_cloudfront_key_value_store" "test" {
  name = var.rName
}
