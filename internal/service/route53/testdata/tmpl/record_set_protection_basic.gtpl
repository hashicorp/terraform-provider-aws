resource "aws_route53_record_set_protection" "test" {
  zone_id     = aws_route53_zone.test.zone_id
  record_name = var.recordName
  record_type = "A"
}

resource "aws_route53_zone" "test" {
  name = var.zoneName
}
