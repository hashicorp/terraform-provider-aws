---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_scraper"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Scraper.
---

# Resource: aws_prometheus_scraper

-> **Note:** You cannot update a scraper. If you change any attribute, Terraform
will delete the current and create a new one.

Provides an Amazon Managed Service for Prometheus fully managed collector
(scraper).

Read more in the [Amazon Managed Service for Prometheus user guide](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-collector.html).


## Example Usage

### Basic Usage

```terraform
resource "aws_prometheus_scraper" "example" {

  destination {
    aws_prometheus_workspace_arn = aws_prometheus_workspace.example.arn
  }

	scrape_configuration = <<EOT
global:
  scrape_interval: 30s
scrape_configs:
  # pod metrics
  - job_name: pod_exporter
    kubernetes_sd_configs:
      - role: pod
  # container metrics
  - job_name: cadvisor
    scheme: https
    authorization:
      credentials_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    kubernetes_sd_configs:
      - role: node
    relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - replacement: kubernetes.default.svc:443
        target_label: __address__
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/$1/proxy/metrics/cadvisor
  # apiserver metrics
  - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    job_name: kubernetes-apiservers
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - action: keep
      regex: default;kubernetes;https
      source_labels:
      - __meta_kubernetes_namespace
      - __meta_kubernetes_service_name
      - __meta_kubernetes_endpoint_port_name
    scheme: https
  # kube proxy metrics
  - job_name: kube-proxy
    honor_labels: true
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    - action: keep
      source_labels:
      - __meta_kubernetes_namespace
      - __meta_kubernetes_pod_name
      separator: '/'
      regex: 'kube-system/kube-proxy.+'
    - source_labels:
      - __address__
      action: replace
      target_label: __address__
      regex: (.+?)(\\:\\d+)?
      replacement: $1:10249
EOT

  source {
    eks_cluster_arn       = data.aws_eks_cluster.example.arn
    subnet_ids            = data.aws_eks_cluster.example.vpc_config[0].subnet_ids
  }
}
```

### Ignoring changes to Prometheus Workspace destination

A managed scraper will add a `AMPAgentlessScraper` tag to its Prometheus workspace
destination. To avoid Terraform state forcing removing the tag from the workspace,
you can add this tag to the destination workspace (preferred) or ignore tags
changes with `lifecycle`. See example below.

```terraform
data "aws_eks_cluster" "this" {
  name = "example"
}

resource "aws_prometheus_workspace" "example" {
  tags = {
		AMPAgentlessScraper = ""
	}
}

resource "aws_prometheus_scraper" "example" {

  source {
    eks_cluster_arn       = data.aws_eks_cluster.example.arn
    subnet_ids            = data.aws_eks_cluster.example.vpc_config[0].subnet_ids
  }

  scrape_configuration = "..."

  destination {
    aws_prometheus_workspace_arn = aws_prometheus_workspace.example.arn
  }
}
```

### Configure aws-auth

Your source Amazon EKS cluster must be configured to allow the scraper to access
metrics. Follow the [user guide](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-collector-how-to.html#AMP-collector-eks-setup)
to setup the appropriate Kubernetes permissions.


## Argument Reference

The following arguments are required:

* `destination` - (Required) Configuration block for the Amazon Managed Service for Prometheus workspace to send metrics to.
* `scrape_configuration` - (Required) The configuration file to use in the new scraper. For more information, see [Scraper configuration](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-collector-how-to.html#AMP-collector-configuration).
* `source` - (Required) Configuration block to specify the Amazon EKS cluster the scraper will collect metrics from.

The following arguments are optional:

* `alias` - (Optional) a name to associate with the scraper. This is for your use, and does not need to be unique.


### destination Arguments

* `aws_prometheus_workspace_arn` - The Amazon Resource Name (ARN) of the prometheus workspace

### source Arguments

* `eks_cluster_arn` - The Amazon Resource Name (ARN) of the source EKS cluster
* `subnet_ids` - List of subnet IDs. Must be in at least two different availability zones.
* `security_group_ids` - List of the security group IDs for the Amazon EKS cluster VPC configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the new scraper.
* `status` - Status of the scraper. One of ACTIVE, CREATING, DELETING, CREATION_FAILED, DELETION_FAILED

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the Managed Scraper using the scraper
identifier. For example:

```terraform
import {
  to = aws_prometheus_scraper.example
  id = "s-0123abc-0000-0123-a000-000000000000"
}
```

Using `terraform import`, import the Managed Scraper using its identifier.
For example:

```console
% terraform import aws_prometheus_scraper.example s-0123abc-0000-0123-a000-000000000000
```
