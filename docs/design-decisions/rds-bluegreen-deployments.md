# RDS Blue/Green Deployments

**Summary:** Discussion of which resource types can support RDS Blue/Green Deployments for updates<br>
**Created:** 2023-11-24<br>
**Updated:** 2024-01-24

---

The Terraform Provider for AWS currently supports [RDS Blue/Green Deployments](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/blue-green-deployments-overview.html) for the resource type [`aws_db_instance`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_instance).
Extending this support to other RDS resource types is a common request.

Due to Terraform's resource model and how RDS implements Blue/Green Deployments, support for Blue/Green Deployments cannot be extended to other resource types.

## Background

Terraform treats each resource as a persistent, self-contained object that can be created, modified, and deleted without interacting with other objects, other than having parameter value dependencies on other resources.
In most cases, this more or less corresponds to how objects within AWS interrelate.

RDS Blue/Green Deployments are implemented using a temporary Blue/Green Deployment orchestration object, which manages both the old and new RDS Instances, data synchronization, switchover, and deletion of the old Instances. The orchestration object is typically deleted after switchover, though it can be retained to, for example, review the operation logs for the deployment.

## Implementations

### `aws_db_instance`

Because the `aws_db_instance` resource type represents a single, self-contained object, we are able to fit within the Terraform resource model while using an RDS Blue/Green Deployment to reduce downtime for many Instance updates.
For instance, changing the backing EC2 instance type or the engine version in place can cause downtimes of over 15 minutes.
When the `blue_green_update.enabled` parameter is set, the AWS Provider will use a Blue/Green Deployment internally to perform the update, so that the downtime is only the length of the switchover.
According to [AWS documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/blue-green-deployments-overview.html), "switchover typically takes under a minute".

The update takes place in a single `terraform apply`, and the Blue/Green Deployment object is not exposed by the provider, since it is only an implementation detail.

### `aws_db_instance` Replicas

Creating a replica RDS DB Instance involves one source `aws_db_instance` resource and one or more replica `aws_db_instance` resources.
Because multiple resources are involved, Terraform cannot treat the source Instance and its replicas as a single unit.
Therefore, Terraform cannot use a Blue/Green Deployment for updating an Instance and its replicas.

### `aws_rds_cluster`

The `aws_rds_cluster` resource type is used to model two types of RDS clusters, either an Aurora cluster or a multi-AZ cluster for a "traditional" database engine such as MySQL or PostgreSQL.

Blue/Green Deployments are not currently supported for multi-AZ clusters.
If, at some point, Blue/Green Deployments are supported, multi-AZ clusters may be a candidate for updates using Blue/Green Deployment for reduced downtime, since the multi-AZ cluster is treated as a single object by both the provider and the AWS API.
However, the multi-AZ cluster already performs a rolling update of the individual instances, so there may be no benefit to using a Blue/Green Deployment for updates.

Aurora clusters do support Blue/Green Deployments.
However, neither the AWS APIs nor the provider treats an Aurora cluster as a self-contained object:
A cluster consists of a containing `aws_rds_cluster` resource and one or more `aws_rds_cluster_instance` resources.
Because of this, a Blue/Green Deployment cannot be used for most updates on an Aurora cluster.

### Standalone Blue/Green Deployment Resource

The RDS Blue/Green Deployment object is a temporary orchestration object which reserves compute and other resources and manages connections.
When it is created, it first creates replicas of any existing RDS Instances, configures synchronization, and updates the engine version, DB parameter group, and instance class if needed.
Once these operations are complete, any additional updates are performed on the new (or "Green") instances.
After these updates are complete, the Green instances can be promoted to be the live instances.

Using a standalone Blue/Green Deployment resource would require multiple iterations of editing the Terraform configuration, applying the configuration, or importing resources.
This isn't a good fit for Terraform.

The following example shows the steps that would be needed to use a standalone Blue/Green Deployment resource with a single RDS Instance (`aws_db_instance`).
Using it with an Aurora cluster would be similar, but require changes to all of the resources making up the cluster.

We start with an RDS Instance with the name `example-db`.

```terraform
resource "aws_db_instance" "example" {
  identifier = "example-db"
}
```

1. Edit the Terraform configuration to add the `aws_rds_blue_green_deployment` resource

    ```terraform
    resource "aws_db_instance" "example" {
      identifier = "example-db"
    }

    resource "aws_rds_blue_green_deployment" "example" {
      source = aws_db_instance.example.arn
    }
    ```

1. Run `terraform apply`.
  This will wait until the Blue/Green Deployment and the "Green" RDS Instance  `example-db-green` are created
1. Edit the Terraform configuration to add an `aws_db_instance` resource for the RDS Instance `example-db-green` and make any desired configuration changes

    ```terraform
    resource "aws_db_instance" "example" {
      identifier = "example-db"
    }

    resource "aws_rds_blue_green_deployment" "example" {
      source = aws_db_instance.example.arn
    }

    resource "aws_db_instance" "example_updated" {
      identifier = "example-db-green"
    }
    ```

1. Run `terraform` to import the RDS Instance `example-db-green`
1. Run `terraform apply` to update the RDS Instance `example-db-green`
1. Edit the Terraform configuration to trigger the switchover on the `aws_rds_blue_green_deployment` resource

    ```terraform
    resource "aws_db_instance" "example" {
      identifier = "example-db"
    }

    resource "aws_rds_blue_green_deployment" "example" {
      source      = aws_db_instance.example.arn
      switch_over = true
    }

    resource "aws_db_instance" "example_updated" {
      identifier = "example-db-green"
    }
    ```

1. Run `terraform apply` to perform the switchover.
  The updated RDS Instance will be renamed from `example-db-green` to `example-db` and the original RDS Instance will be renamed from `example-db` to `example-db-old`.
  This means that the resource `aws_db_instance.example` will now be pointing to the updated RDS Instance, the resource `aws_db_instance.example_updated` will be pointing at a non-existent RDS Instance, and the original RDS Instance is not known to Terraform.
1. Edit the Terraform configuration to remove the resource `aws_db_instance.example_updated` and add an `aws_db_instance` resource for the RDS Instance `example-db-old`

    ```terraform
    resource "aws_db_instance" "example" {
      identifier = "example-db"
    }

    resource "aws_rds_blue_green_deployment" "example" {
      source      = aws_db_instance.example.arn
      switch_over = true
    }

    resource "aws_db_instance" "example_to_delete" {
      identifier = "example-db-old"
    }
    ```

1. Run `terraform` to import the RDS Instance `example-db-old`
1. Edit the Terraform configuration to remove the resource `aws_db_instance.example_updated` and the `aws_rds_blue_green_deployment` resource

    ```terraform
    resource "aws_db_instance" "example" {
      identifier = "example-db"
    }
    ```

1. Run `terraform apply` to delete the Blue/Green Deployment and the original RDS Instance
