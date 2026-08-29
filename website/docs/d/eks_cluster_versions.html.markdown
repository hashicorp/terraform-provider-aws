---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_cluster_versions"
description: |-
  Terraform data source for managing AWS EKS (Elastic Kubernetes) Cluster Versions.
---

# Data Source: aws_eks_cluster_versions

Terraform data source for managing AWS EKS (Elastic Kubernetes) Cluster Versions.

## Example Usage

### Basic Usage

```terraform
data "aws_eks_cluster_versions" "example" {}

output "eks_cluster_versions" {
  value = data.aws_eks_cluster_versions.example.cluster_versions
}

output "eks_cluster_version_filtered" {
  value = [for version in data.aws_eks_cluster_versions.example.cluster_versions : version if version.cluster_version == "1.33"]
}

output "eks_cluster_version_list" {
  value = [for version in data.aws_eks_cluster_versions.example.cluster_versions : version.cluster_version]
}
```

### Filter by Cluster Type

```terraform
data "aws_eks_cluster_versions" "example" {
  cluster_type = "eks"
}
```

### Filter by Version Status

```terraform
data "aws_eks_cluster_versions" "example" {
  version_status = "STANDARD_SUPPORT"
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_type` - (Optional) Type of clusters to filter by.
Currently, the only valid value is `eks`.
* `default_only` - (Optional) Whether to show only the default versions of Kubernetes supported by EKS.
* `include_all` - (Optional) Whether to include all kubernetes versions in the response.
* `version_status` - (Optional) Status of the EKS cluster versions to list.
Valid values are `STANDARD_SUPPORT` or `UNSUPPORTED` or `EXTENDED_SUPPORT`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cluster_versions` - A list of Kubernetes version information.
    * `cluster_type` - Type of cluster that the version belongs to.
    * `cluster_version` - Kubernetes version supported by EKS.
    * `control_plane_component_config` - Default control plane component configuration and constraints for this version.
        * `kube_api_server_config` - Kubernetes API server configuration defaults and constraints.
            * `event_ttl` - Event TTL configuration with default value and constraints.
                * `default_value` - The default value for the event TTL.
                * `constraints` - The constraints for the event TTL.
                    * `min` - The minimum allowed duration.
                    * `max` - The maximum allowed duration.
            * `service_node_port_range` - Service node port range configuration with default value and constraints.
                * `default_value` - The default port range.
                    * `min_port` - The default minimum port.
                    * `max_port` - The default maximum port.
                * `constraints` - The constraints for the port range.
                    * `min_port` - The allowed range for the minimum port (`min`, `max`).
                    * `max_port` - The allowed range for the maximum port (`min`, `max`).
        * `kube_controller_manager_config` - Kubernetes controller manager configuration defaults and constraints.
            * `horizontal_pod_autoscaler_controller_config` - HPA controller configuration defaults and constraints.
                * `horizontal_pod_autoscaler_sync_period` - HPA sync period configuration with default value and constraints.
                    * `default_value` - The default sync period.
                    * `constraints` - The constraints for the sync period (`min`, `max`).
            * `pod_gc_controller_config` - Pod garbage collection controller configuration defaults and constraints.
                * `terminated_pod_gc_threshold` - Terminated pod GC threshold configuration with default value and constraints.
                    * `default_value` - The default terminated pod GC threshold.
                    * `constraints` - The constraints for the terminated pod GC threshold (`min`, `max`).
        * `kube_scheduler_config` - Kubernetes scheduler configuration defaults and constraints.
            * `node_resources_fit` - NodeResourcesFit plugin configuration with default value and constraints.
                * `scoring_strategy` - Scoring strategy configuration.
                    * `default_value` - Default scoring strategy (`type`, `resources`).
                    * `constraints` - Scoring strategy constraints.
                        * `scoring_strategy` - Allowed values for the strategy type.
                        * `resources` - Constraints for resource names and weights.
    * `control_plane_scaling_tiers` - Available provisioned control plane scaling tiers and their capabilities.
        * `tier_name` - The name of the scaling tier.
        * `api_request_concurrency` - Maximum API request concurrency supported by this tier.
        * `pod_scheduling_rate_per_second` - Maximum pod scheduling rate per second supported by this tier.
        * `cluster_database_size_gb` - Maximum cluster database size in GB supported by this tier.
        * `control_plane_component_config_overrides` - Control plane component configuration overrides specific to this tier (same structure as `control_plane_component_config`).
    * `default_platform_version` - Default eks platform version for the cluster version.
    * `default_version` - Default Kubernetes version for the cluster version.
    * `end_of_extended_support_date` - End of extended support date for the cluster version.
    * `end_of_standard_support_date` - End of standard support date for the cluster version.
    * `kubernetes_patch_version` - Kubernetes patch version for the cluster version.
    * `release_date` - Release date of the cluster version.
    * `version_status` - Status of the EKS cluster version.
