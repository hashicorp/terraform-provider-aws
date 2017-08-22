---
layout: "aws"
page_title: "AWS: aws_emr_instance_fleet"
sidebar_current: "docs-aws-resource-emr-instance-fleet"
description: |-
  Provides an Elastic MapReduce Cluster Instance Fleet
---

# aws_emr_instance_fleet

Provides an Elastic MapReduce Cluster Instance Fleet configuration.
See [Amazon Elastic MapReduce Documentation](https://aws.amazon.com/documentation/emr/) for more information.

~> **NOTE:** At this time, Instance Fleets cannot be destroyed through the API nor
web interface. Instance Fleets are destroyed when the EMR Cluster is destroyed.
Terraform will resize any Instance Fleet to zero when destroying the resource.

## Example Usage

```hcl
resource "aws_emr_instance_fleet" "task" {
    cluster_id            = "${aws_emr_cluster.tf-test-cluster.id}"
    instance_fleet_type   = "TASK"
    instance_type_configs [
        {
            bid_price_as_percentage_of_on_demand_price = 100
            configurations = []
            ebs_optimized = true
            ebs_config = [
                {
                    iops  = 300
                    size  = 10
                    type  = "gp2"
                    volumes_per_instance = 1
                }
            ]

            instance_type       = "m3.xlarge"
            "weighted_capacity" = 8
        }
    ],
    launch_specifications {
        spot_specification {
            block_duration_minutes   = 60
            timeout_action           = "TERMINATE_CLUSTER"
            timeout_duration_minutes = 10
        }
    },
    name                      = "my little instance fleet"
    target_on_demand_capacity = 1
    target_spot_capacity      = 1
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) ID of the EMR Cluster to attach to. Changing this forces a new resource to be created.

* `instance_fleet_type` - (Required) The node type that the instance fleet hosts. Valid values are `MASTER`, `CORE`, and `TASK`. Changing this forces a new resource to be created.

* `instance_type_configs` - (Optional) The instance type configurations that define the EC2 instances in the instance fleet. List of `instance_type_config` blocks. 

* `launch_specifications` - (Optional) The launch specification for the instance fleet. 
    
    * `spot_specification` - (Required) The launch specification for Spot instances in the fleet, which determines the 
    defined duration and provisioning timeout behavior.

* `name` - (Optional) The friendly name of the instance fleet.

* `target_on_demand_capacity` - (Optional) The target capacity of On-Demand units for the instance fleet, which determines how many On-Demand instances to provision.

* `target_spot_capacity` - (Optional) The target capacity of Spot units for the instance fleet, which determines how many Spot instances to provision.



`instance_type_config` supports the following:

* `bid_price` - (Optional) The bid price for each EC2 Spot instance type as defined by `instance_type`. 
Expressed in USD. If neither `bid_price` nor `bid_price_as_percentage_of_on_demand_price` is provided, 
`bid_price_as_percentage_of_on_demand_price` defaults to 100%.

* `bid_price_as_percentage_of_on_demand_price` - (Optional) The bid price, as a percentage of On-Demand price, 
for each EC2 Spot instance as defined by `instance_type`. Expressed as a number (for example, 20 specifies 20%). 
If neither `bid_price` nor `bid_price_as_percentage_of_on_demand_price` is provided, 
`bid_price_as_percentage_of_on_demand_price` defaults to 100%.

* `configurations` - (Optional) A configuration classification that applies when provisioning cluster instances, 
which can include configurations for applications and software that run on the cluster. List of `configuration` blocks.

* `ebs_optimized` - (Optional) Indicates whether an Amazon EBS volume is EBS-optimized.

* `ebs_config` - (Optional) The configuration of Amazon Elastic Block Storage (EBS) attached to each instance as
defined by `instance_type`.

* `instance_type` - (Required) An EC2 instance type, such as m3.xlarge.

* `weighted_capacity` - (Optional) The number of units that a provisioned instance of this type provides toward 
fulfilling the target capacities defined in `aws_emr_instance_fleet`. This value is 1 for a master instance fleet, 
and must be 1 or greater for core and task instance fleets. Defaults to 1 if not specified.



`spot_specification` supports the following:

* `block_duration_minutes` - (Optional) The defined duration for Spot instances (also known as Spot blocks) in minutes. 
When specified, the Spot instance does not terminate before the defined duration expires, and defined duration pricing 
for Spot instances applies. Valid values are 60, 120, 180, 240, 300, or 360. The duration period starts as soon as a 
Spot instance receives its instance ID. At the end of the duration, Amazon EC2 marks the Spot instance for termination 
and provides a Spot instance termination notice, which gives the instance a two-minute warning before it terminates.

* `timeout_action` - (Required) The action to take when `target_spot_capacity` has not been fulfilled when the 
`timeout_duration_minutes` has expired. Spot instances are not uprovisioned within the Spot provisioining timeout.
Valid values are `TERMINATE_CLUSTER` and `SWITCH_TO_ON_DEMAND`. `SWITCH_TO_ON_DEMAND` specifies that if no Spot 
instances are available, On-Demand Instances should be provisioned to fulfill any remaining Spot capacity.

* `timeout_duration_minutes` - (Required) The spot provisioning timeout period in minutes. If Spot instances are not 
provisioned within this time period, the `timeout_action` is taken. Minimum value is 5 and maximum value is 1440. 
The timeout applies only during initial provisioning, when the cluster is first created.



`configuration` supports the following:

* `classification` - (Optional) The classification within a configuration.

* `configurations` - (Optional) A list of additional configurations to apply within a configuration object.

* `properties` - (Optional) A set of properties specified within a configuration classification.



`ebs_config` EBS volume specifications such as volume type, IOPS, and size (GiB) that will be requested for the EBS volume attached to an EC2 instance in the cluster.

* `iops` - (Optional) The number of I/O operations per second (IOPS) that the volume supports.
    
* `size_in_gb` - (Required) The volume size, in gibibytes (GiB). This can be a number from 1 - 1024. If the volume type is EBS-optimized, the minimum value is 10.
    
* `volume_type` - (Required) The volume type. Volume types supported are `gp2`, `io1`, `standard`.

* `volumes_per_instance` - (Optional) Number of EBS volumes with a specific volume configuration that will be associated with every instance in the instance group


## Attributes Reference

The following attributes are exported:

* `id` - The unique identifier of the instance fleet.

* `provisioned_on_demand_capacity` The number of On-Demand units that have been provisioned for the instance 
fleet to fulfill TargetOnDemandCapacity. This provisioned capacity might be less than or greater than TargetOnDemandCapacity.

* `provisioned_spot_capacity` The number of Spot units that have been provisioned for this instance fleet 
to fulfill TargetSpotCapacity. This provisioned capacity might be less than or greater than TargetSpotCapacity.

* `status` The current status of the instance fleet.
