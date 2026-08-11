resource "aws_auditmanager_control" "test" {
{{- template "region" }}
  name = var.rName

  control_mapping_sources {
    source_name          = var.rName
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }

{{ template "tags" . }}
}
