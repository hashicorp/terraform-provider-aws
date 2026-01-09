---
subcategory: "{{ .HumanFriendlyService }}"
layout: "aws"
page_title: "AWS: aws_{{ .ServicePackage }}_{{ .ListResourceSnake }}"
description: |-
  Lists {{ .HumanFriendlyService }} {{ .HumanResourceName }} resources.
---

# List Resource: aws_{{ .ServicePackage }}_{{ .ListResourceSnake }}

Lists {{ .HumanFriendlyService }} {{ .HumanResourceName }} resources.

## Example Usage

```terraform
list "aws_{{ .ServicePackage }}_{{ .ListResourceSnake }}" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.