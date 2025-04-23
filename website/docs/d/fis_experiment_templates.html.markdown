---
subcategory: "FIS (Fault Injection Simulator)"
layout: "aws"
page_title: "AWS: aws_fis_experiment_templates"
description: |-
  Get information about a set of FIS Experiment Templates
---

# Data Source: aws_fis_experiment_templates

This resource can be useful for getting back a set of FIS experiment template IDs.

## Example Usage

The following shows outputting a list of all FIS experiment template IDs

```terraform
data "aws_fis_experiment_templates" "all" {}

output "all" {
  value = data.aws_fis_experiment_templates.all.ids
}
```

The following shows filtering FIS experiment templates by tag

```terraform
data "aws_fis_experiment_templates" "example" {
  tags = {
    Name = "example"
    Tier = 1
  }
}

data "aws_iam_policy_document" "example" {
  statement {
    sid     = "StartFISExperiment"
    effect  = "Allow"
    actions = ["fis:StartExperiment"]
    resources = [
      "arn:aws:fis:*:*:experiment-template/${data.aws_fis_experiment_templates.example.ids[0]}",
      "arn:aws:fis:*:*:experiment/*"
    ]
  }
}
```

## Argument Reference

* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired experiment templates.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of all the experiment template ids found.
