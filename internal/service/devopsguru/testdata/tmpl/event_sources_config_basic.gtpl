resource "aws_devopsguru_event_sources_config" "test" {
{{- template "region" }}
  event_sources {
    amazon_code_guru_profiler {
      status = "ENABLED"
    }
  }
}
