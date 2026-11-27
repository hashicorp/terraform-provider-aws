---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_anomaly_detector"
description: |-
  Lists AMP (Managed Prometheus) Anomaly Detector resources.
---

# List Resource: aws_prometheus_anomaly_detector

Lists AMP (Managed Prometheus) Anomaly Detector resources.

## Example Usage

```terraform
list "aws_prometheus_anomaly_detector" "example" {
  provider = aws

  config {
    workspace_id = aws_prometheus_workspace.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `alias` - (Optional) Alias of the Anomaly Detector to filter results by.
* `region` - (Optional) Region to query. Defaults to provider region.
* `workspace_id` - (Required) ID of the AMP workspace to list Anomaly Detectors from.
