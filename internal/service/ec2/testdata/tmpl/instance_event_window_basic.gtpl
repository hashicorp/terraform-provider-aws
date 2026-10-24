resource "aws_ec2_instance_event_window" "test" {
{{- template "region" }}
  name = var.rName

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }

{{- template "tags" . }}
}
