---
subcategory: "{{ .HumanFriendlyService }}"
layout: "aws"
page_title: "AWS: aws_{{ .ServicePackage }}_{{ .ResourceSnake }}"
description: |-
  Lists {{ .HumanFriendlyService }} {{ .HumanResourceName }} resources.
---

# List Resource: aws_{{ .ServicePackage }}_{{ .ResourceSnake }}

Lists {{ .HumanFriendlyService }} {{ .HumanResourceName }} resources.

## Example Usage

```terraform
list "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.