---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_scraper"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Scraper.
---

# Resource: aws_prometheus_scraper

-> **Note:** If you change a Scraper's source (EKS cluster), Terraform
will delete the current Scraper and create a new one.

Provides an Amazon Managed Service for Prometheus fully managed collector
(scraper).

Read more in the [Amazon Managed Service for Prometheus user guide](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-collector.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_prometheus_scraper" "example" {
  source {
    eks {
      cluster_arn = data.aws_eks_cluster.example.arn
      subnet_ids  = data.aws_eks_cluster.example.vpc_config[0].subnet_ids
    }
  }

  destination {
    amp {
      workspace_arn = aws_prometheus_workspace.example.arn
    }
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
}
```

### Use default EKS scraper configuration

You can use the data source `aws_prometheus_scraper_configuration` to use a
service managed scrape configuration.

```terraform
data "aws_prometheus_default_scraper_configuration" "example" {}

resource "aws_prometheus_scraper" "example" {

  destination {
    amp {
      workspace_arn = aws_prometheus_workspace.example.arn
    }
  }

  scrape_configuration = data.aws_prometheus_scraper_configuration.example.configuration

  source {
    eks {
      cluster_arn = data.aws_eks_cluster.example.arn
      subnet_ids  = data.aws_eks_cluster.example.vpc_config[0].subnet_ids
    }
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
    eks {
      cluster_arn = data.aws_eks_cluster.example.arn
      subnet_ids  = data.aws_eks_cluster.example.vpc_config[0].subnet_ids
    }
  }

  scrape_configuration = "..."

  destination {
    amp {
      workspace_arn = aws_prometheus_workspace.example.arn
    }
  }
}
```

### Configure aws-auth

Your source Amazon EKS cluster must be configured to allow the scraper to access
metrics. Follow the [user guide](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-collector-how-to.html#AMP-collector-eks-setup)
to setup the appropriate Kubernetes permissions.

### Cross-Account Configuration

This setup allows the scraper, running in a source account, to remote write its collected metrics to a workspace in a target account. Note that:

- The target Role and target Workspace must be in the same account
- The source Scraper and target Workspace must be in the same Region

Follow [the AWS Best Practices guide](https://aws-observability.github.io/observability-best-practices/patterns/ampxa) to learn about the IAM roles configuration and overall setup.

```terraform
resource "aws_prometheus_scraper" "example" {
  source {
    eks {
      cluster_arn = data.aws_eks_cluster.example.arn
      subnet_ids  = data.aws_eks_cluster.example.vpc_config[0].subnet_ids
    }
  }

  destination {
    amp {
      workspace_arn = "<target_account_workspace_arn>"
    }
  }

  role_configuration {
    source_role_arn = aws_iam_role.source.arn
    target_role_arn = "arn:aws:iam::ACCOUNT-ID:role/target-role-name"
  }

  scrape_configuration = "..."
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `destination` - (Required) Configuration block for the managed scraper to send metrics to. See [`destination`](#destination).
* `scrape_configuration` - (Required) The configuration file to use in the new scraper. For more information, see [Scraper configuration](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-collector-how-to.html#AMP-collector-configuration).
* `source` - (Required) Configuration block to specify where the managed scraper will collect metrics from. See [`source`](#source).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `alias` - (Optional) a name to associate with the managed scraper. This is for your use, and does not need to be unique.

* `role_configuration` - (Optional) Configuration block to enable writing to an Amazon Managed Service for Prometheus workspace in a different account. See [`role_configuration`](#role_configuration) below.

### `destination`

* `amp` - (Required) Configuration block for an Amazon Managed Prometheus workspace destination. See [`amp`](#amp).

### `amp`

* `workspace_arn` - (Required) The Amazon Resource Name (ARN) of the prometheus workspace.

### `source`

* `eks` - (Required) Configuration block for an EKS cluster source. See [`eks`](#eks).

#### `eks`

* `eks_cluster_arn` - (Required) The Amazon Resource Name (ARN) of the source EKS cluster.
* `subnet_ids` - (Required) List of subnet IDs. Must be in at least two different availability zones.
* `security_group_ids` - (Optional) List of the security group IDs for the Amazon EKS cluster VPC configuration.

### `role_configuration`

* `source_role_arn` - (Required) The Amazon Resource Name (ARN) of the source role configuration. Must be an IAM role ARN.
* `target_role_arn` - (Required) The Amazon Resource Name (ARN) of the target role configuration. Must be an IAM role ARN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the new scraper.
* `role_arn` - The Amazon Resource Name (ARN) of the IAM role that provides permissions for the scraper to discover, collect, and produce metrics
* `status` - Status of the scraper. One of ACTIVE, CREATING, DELETING, CREATION_FAILED, DELETION_FAILED

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `2m`)
* `delete` - (Default `20m`)

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
