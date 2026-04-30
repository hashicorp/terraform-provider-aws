resource "aws_opensearchserverless_collection_group" "test" {
{{- template "region" }}
  name             = var.rName
  standby_replicas = "ENABLED"

  capacity_limits {
    max_indexing_capacity_in_ocu = 1
    max_search_capacity_in_ocu   = 1
  }

{{- template "tags" . }}
}
