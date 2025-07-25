resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.region]
    }
  }
{{- template "tags" . }}
}

resource "aws_guardduty_detector" "test" {}

data "aws_region" "current" {}
