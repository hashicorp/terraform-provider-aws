resource "aws_iot_event_configurations" "test" {
{{- template "region" }}
  event_configurations = {
    "THING"                  = true,
    "THING_GROUP"            = false,
    "THING_TYPE"             = false,
    "THING_GROUP_MEMBERSHIP" = false,
    "THING_GROUP_HIERARCHY"  = false,
    "THING_TYPE_ASSOCIATION" = false,
    "JOB"                    = false,
    "JOB_EXECUTION"          = false,
    "POLICY"                 = false,
    "CERTIFICATE"            = true,
    "CA_CERTIFICATE"         = true,
  }
}
