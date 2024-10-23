resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name

{{- template "tags" . }}
}

resource "aws_resourcegroups_group" "test" {
  name = var.rName

  resource_query {
    query = <<JSON
	{
		"ResourceTypeFilters": [
		  "AWS::EC2::Instance"
		],
		"TagFilters": [
		  {
			"Key": "Stage",
			"Values": [
			  "Test"
			]
		  }
		]
	  }
JSON
  }
}
