resource "aws_guardduty_detector" "test" {
  {{- template "region" }}
  enable = true
}

resource "aws_guardduty_detector_feature" "test" {
  {{- template "region" }}
  detector_id = aws_guardduty_detector.test.id
  name        = "RDS_LOGIN_EVENTS"
  status      = "ENABLED"
}
