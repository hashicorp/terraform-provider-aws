resource "aws_xray_trace_segment_destination" "test" {
{{- template "region" }}
  destination = "XRay"
}
