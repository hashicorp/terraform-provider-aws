package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	//"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsAutoscalingGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAutoscalingGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"default_cool_down": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"desired_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"health_check_grace_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"health_check_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_configuration_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancer_names": {
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

func dataSourceAwsAutoscalingGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	d.SetId(time.Now().UTC().String())

	groupName, _ := d.GetOk("name")

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(groupName.(string)),
		},
	}

	log.Printf("[DEBUG] Reading Autoscaling Group: %s", input)

	result, err := conn.DescribeAutoScalingGroups(input)
	log.Printf("[DEBUG] Checking for error: %s", err)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeInvalidNextToken:
				return fmt.Errorf("%s %s", autoscaling.ErrCodeInvalidNextToken, aerr.Error())
			case autoscaling.ErrCodeResourceContentionFault:
				return fmt.Errorf("%s %s", autoscaling.ErrCodeResourceContentionFault, aerr.Error())
			default:
				return fmt.Errorf("%s", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return fmt.Errorf("%s", err.Error())
		}
	}

	log.Printf("[DEBUG] Found Autoscaling Group: %s", result)

	var group *autoscaling.Group

	if len(result.AutoScalingGroups) < 1 {
		return fmt.Errorf("Your query did not return any results. Please try a different search criteria.")
	}

	if len(result.AutoScalingGroups) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more " +
			"specific search criteria.")
	} else {
		group = result.AutoScalingGroups[0]
	}

	log.Printf("[DEBUG] aws_autoscaling_group - Single Auto Scaling Group found: %s", *group.AutoScalingGroupName)

	if err := groupDescriptionAttributes(d, group); err != nil {
		return err
	}

	return nil
}

// Compile the list of availability zones
func setAvailabilityZones(d *schema.ResourceData, group *autoscaling.Group) error {
	zones := make([]string, 0, len(group.AvailabilityZones))
	for _, zone := range group.AvailabilityZones {
		zones = append(zones, *zone)
	}
	if err := d.Set("availability_zones", zones); err != nil {
		return err
	}

	return nil
}

// Compile the list of load balancers
func setLoadBalancers(d *schema.ResourceData, group *autoscaling.Group) error {
	balancers := make([]string, 0, len(group.LoadBalancerNames))
	for _, lb := range group.LoadBalancerNames {
		balancers = append(balancers, *lb)
	}
	if err := d.Set("load_balancer_names", balancers); err != nil {
		return err
	}

	return nil
}

// Populate group attribute fields with the returned group
func groupDescriptionAttributes(d *schema.ResourceData, group *autoscaling.Group) error {
	log.Printf("[DEBUG] Setting attributes: %s", group)
	d.Set("name", group.AutoScalingGroupName)
	d.Set("arn", group.AutoScalingGroupARN)
	if err := setAvailabilityZones(d, group); err != nil {
		return err
	}
	d.Set("default_cool_down", group.DefaultCooldown)
	d.Set("desired_capacity", group.DesiredCapacity)
	d.Set("health_check_grace_period", group.HealthCheckGracePeriod)
	d.Set("health_check_type", group.HealthCheckType)
	d.Set("launch_configuration_name", group.LaunchConfigurationName)
	if err := setLoadBalancers(d, group); err != nil {
		return err
	}
	d.Set("max_size", group.MaxSize)
	d.Set("min_size", group.MinSize)
	d.Set("new_instances_protected_from_scale_in", group.NewInstancesProtectedFromScaleIn)
	d.Set("placement_group", group.PlacementGroup)
	d.Set("service_linked_role_arn", group.ServiceLinkedRoleARN)
	d.Set("status", group.Status)
	d.Set("target_group_arns", group.TargetGroupARNs)
	d.Set("termination_policies", group.TerminationPolicies)
	d.Set("vpc_zone_identifier", group.VPCZoneIdentifier)

	return nil
}
