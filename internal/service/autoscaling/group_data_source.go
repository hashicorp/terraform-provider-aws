package autoscaling

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_cooldown": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"desired_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"enabled_metrics": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"health_check_grace_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"health_check_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_configuration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_template": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"load_balancers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"max_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"new_instances_protected_from_scale_in": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"placement_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_linked_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_group_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"termination_policies": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vpc_zone_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	groupName := d.Get("name").(string)
	group, err := FindGroupByName(conn, groupName)

	if err != nil {
		return fmt.Errorf("reading Auto Scaling Group (%s): %w", groupName, err)
	}

	d.SetId(aws.StringValue(group.AutoScalingGroupName))
	d.Set("arn", group.AutoScalingGroupARN)
	d.Set("availability_zones", aws.StringValueSlice(group.AvailabilityZones))
	d.Set("default_cooldown", group.DefaultCooldown)
	d.Set("desired_capacity", group.DesiredCapacity)
	d.Set("enabled_metrics", flattenEnabledMetrics(group.EnabledMetrics))
	d.Set("health_check_grace_period", group.HealthCheckGracePeriod)
	d.Set("health_check_type", group.HealthCheckType)
	d.Set("launch_configuration", group.LaunchConfigurationName)
	if group.LaunchTemplate != nil {
		if err := d.Set("launch_template", []interface{}{flattenLaunchTemplateSpecification(group.LaunchTemplate)}); err != nil {
			return fmt.Errorf("setting launch_template: %w", err)
		}
	} else {
		d.Set("launch_template", nil)
	}
	d.Set("load_balancers", aws.StringValueSlice(group.LoadBalancerNames))
	d.Set("max_size", group.MaxSize)
	d.Set("min_size", group.MinSize)
	d.Set("name", group.AutoScalingGroupName)
	d.Set("new_instances_protected_from_scale_in", group.NewInstancesProtectedFromScaleIn)
	d.Set("placement_group", group.PlacementGroup)
	d.Set("service_linked_role_arn", group.ServiceLinkedRoleARN)
	d.Set("status", group.Status)
	d.Set("target_group_arns", aws.StringValueSlice(group.TargetGroupARNs))
	d.Set("termination_policies", aws.StringValueSlice(group.TerminationPolicies))
	d.Set("vpc_zone_identifier", group.VPCZoneIdentifier)

	return nil
}
