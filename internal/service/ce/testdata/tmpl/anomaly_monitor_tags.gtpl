resource "aws_ce_anomaly_monitor" "test" {
{{- template "region" }}
  name         = var.rName
  monitor_type = "CUSTOM"

  monitor_specification = <<JSON
{
	"And": null,
	"CostCategories": null,
	"Dimensions": null,
	"Not": null,
	"Or": null,
	"Tags": {
		"Key": "user:CostCenter",
		"MatchOptions": null,
		"Values": [
			"10000"
		]
	}
}
JSON
{{- template "tags" . }}
}
