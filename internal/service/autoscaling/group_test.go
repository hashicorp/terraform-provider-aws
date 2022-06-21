package autoscaling_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elbv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(autoscaling.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"gp3 is invalid",
	)
}

func testAccGroupImportStep(n string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      n,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"force_delete",
			"initial_lifecycle_hook",
			"tag",
			"tags",
			"wait_for_capacity_timeout",
			"wait_for_elb_capacity",
		},
	}
}

func TestAccAutoScalingGroup_basic(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf(`autoScalingGroup:.+:autoScalingGroupName/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", "false"),
					resource.TestCheckResourceAttr(resourceName, "context", ""),
					resource.TestCheckResourceAttr(resourceName, "default_cooldown", "300"),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "force_delete", "false"),
					resource.TestCheckResourceAttr(resourceName, "force_delete_warm_pool", "false"),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", "aws_launch_configuration.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "metrics_granularity", "1Minute"),
					resource.TestCheckNoResourceAttr(resourceName, "min_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "placement_group", ""),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", "false"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_linked_role_arn", "iam", "role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "tag.#"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_capacity_timeout", "10m"),
					resource.TestCheckNoResourceAttr(resourceName, "wait_for_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", "0"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_disappears(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					acctest.CheckResourceDisappears(acctest.Provider, tfautoscaling.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingGroup_nameGenerated(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_namePrefix(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_tags(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_tags1(rName, "key1", "value1", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":                 "key1",
						"value":               "value1",
						"propagate_at_launch": "true",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_tags2(rName, "key1", "value1updated", true, "key2", "value2", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":                 "key1",
						"value":               "value1updated",
						"propagate_at_launch": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":                 "key2",
						"value":               "value2",
						"propagate_at_launch": "false",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_tags1(rName, "key2", "value2", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":                 "key2",
						"value":               "value2",
						"propagate_at_launch": "true",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_deprecatedTags(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_deprecatedTags1(rName, "key1", "value1", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tags.*", map[string]string{
						"key":                 "key1",
						"value":               "value1",
						"propagate_at_launch": "true",
					}),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_simple(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_simple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf(`autoScalingGroup:.+:autoScalingGroupName/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cooldown", "300"),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", "4"),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "force_delete", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_delete_warm_pool", "false"),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", "aws_launch_configuration.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "metrics_granularity", "1Minute"),
					resource.TestCheckNoResourceAttr(resourceName, "min_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "placement_group", ""),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", "false"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_linked_role_arn", "iam", "role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":                 "Name",
						"value":               rName,
						"propagate_at_launch": "true",
					}),
					resource.TestCheckNoResourceAttr(resourceName, "tags.#"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "OldestInstance"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.1", "ClosestToNextInstanceHour"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_capacity_timeout", "10m"),
					resource.TestCheckNoResourceAttr(resourceName, "wait_for_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", "0"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_simpleUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf(`autoScalingGroup:.+:autoScalingGroupName/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_cooldown", "300"),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", "4"),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "force_delete", "true"),
					resource.TestCheckResourceAttr(resourceName, "force_delete_warm_pool", "false"),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "400"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", "aws_launch_configuration.test2", "name"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "6"),
					resource.TestCheckResourceAttr(resourceName, "metrics_granularity", "1Minute"),
					resource.TestCheckNoResourceAttr(resourceName, "min_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "3"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "placement_group", ""),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", "true"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_linked_role_arn", "iam", "role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						"key":                 "Name",
						"value":               rName,
						"propagate_at_launch": "true",
					}),
					resource.TestCheckNoResourceAttr(resourceName, "tags.#"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "ClosestToNextInstanceHour"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_capacity_timeout", "10m"),
					resource.TestCheckNoResourceAttr(resourceName, "wait_for_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", "0"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_terminationPolicies(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_terminationPoliciesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "OldestInstance"),
				),
			},
			{
				Config: testAccGroupConfig_terminationPoliciesExplicitDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "Default"),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "0"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_vpcUpdates(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_az(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_vpcZoneIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_zone_identifier.*", "aws_subnet.test.0", "id"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withLoadBalancer(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_loadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "force_delete", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "load_balancers.*", "aws_elb.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_elb_capacity", "2"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_WithLoadBalancer_toTargetGroup(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_loadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_target2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "1"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_loadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withPlacementGroup(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_placement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "placement_group", "aws_placement_group.test", "name"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_enablingMetrics(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_enabledMetrics(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTotalInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupPendingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTerminatingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupDesiredCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupMaxSize"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withMetrics(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_allMetricsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "13"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTotalInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupPendingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTerminatingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupDesiredCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupInServiceInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupMinSize"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupMaxSize"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupPendingCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupInServiceCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupStandbyCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTotalCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTerminatingCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupStandbyInstances"),
				),
			},
			{
				Config: testAccGroupConfig_enabledMetrics(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTotalInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupPendingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTerminatingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupDesiredCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupMaxSize"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_suspendingProcesses(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_suspendedProcesses(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "suspended_processes.*", "AlarmNotification"),
					resource.TestCheckTypeSetElemAttr(resourceName, "suspended_processes.*", "ScheduledActions"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_suspendedProcessesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "suspended_processes.*", "AZRebalance"),
					resource.TestCheckTypeSetElemAttr(resourceName, "suspended_processes.*", "ScheduledActions"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_serviceLinkedRoleARN(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_serviceLinkedRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "service_linked_role_arn", "data.aws_iam_role.test", "arn"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_maxInstanceLifetime(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_maxInstanceLifetime(rName, 864000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", "864000"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_maxInstanceLifetime(rName, 604800),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", "604800"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_initialLifecycleHook(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_initialLifecycleHook(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 2),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "initial_lifecycle_hook.*", map[string]string{
						"default_result":       "CONTINUE",
						"heartbeat_timeout":    "30",
						"lifecycle_transition": "autoscaling:EC2_INSTANCE_LAUNCHING",
						"name":                 "launching",
					}),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_launchTemplate(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_launchTemplate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", "aws_launch_template.test", "default_version"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_LaunchTemplate_update(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_launchTemplate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", "aws_launch_template.test", "default_version"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", "aws_launch_configuration.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_launchTemplateName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", ""),
				),
			},
			{
				Config: testAccGroupConfig_launchTemplateLatestVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Latest"),
				),
			},
			{
				Config: testAccGroupConfig_launchTemplate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", "aws_launch_template.test", "default_version"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_largeDesiredCapacity(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_largeDesiredCapacity(rName, 101),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 101),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", "101"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "101"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "101"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_basic(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceRefreshBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshMinHealthyPercentage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshSkipMatching(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshFull(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", "25"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.0", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.1", "20"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.2", "25"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.3", "50"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.4", "100"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", "10"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "0"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_start(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"
	launchConfigurationResourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceRefreshStart(rName, acctest.ResourcePrefix+"-1-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationResourceName, "name"),
					testAccCheckInstanceRefreshCount(&group, 0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshStart(rName, acctest.ResourcePrefix+"-2-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationResourceName, "name"),
					testAccCheckInstanceRefreshCount(&group, 1),
					testAccCheckInstanceRefreshStatus(&group, 0, autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshStart(rName, acctest.ResourcePrefix+"-3-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationResourceName, "name"),
					testAccCheckInstanceRefreshCount(&group, 2),
					testAccCheckInstanceRefreshStatus(&group, 0, autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress),
					testAccCheckInstanceRefreshStatus(&group, 1, autoscaling.InstanceRefreshStatusCancelled),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_triggers(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceRefreshBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshTriggers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_refresh.0.triggers.*", "tags"),
					testAccCheckInstanceRefreshCount(&group, 1),
					testAccCheckInstanceRefreshStatus(&group, 0, autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/256
func TestAccAutoScalingGroup_loadBalancers(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_loadBalancers(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "11"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_loadBalancers(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_loadBalancers(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "11"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_targetGroups(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_target(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_target(rName, 12),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "12"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_target(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_ALBTargetGroups_elbCapacity(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var tg elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_targetELBCapacity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					testAccCheckLBTargetGroupExists("aws_lb_target_group.test", &tg),
					testAccCheckALBTargetGroupHealthy(&tg),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_warmPool(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_warmPoolEmpty(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.max_group_prepared_capacity", "-1"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.min_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.pool_state", "Stopped"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_warmPoolFull(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.0.reuse_on_scale_in", "true"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.max_group_prepared_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.min_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.pool_state", "Stopped"),
				),
			},
			{
				Config: testAccGroupConfig_warmPoolNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckNoResourceAttr(resourceName, "warm_pool.#"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_launchTempPartitionNum(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_partition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_Destroy_whenProtectedFromScaleIn(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_destroyWhenProtectedFromScaleInBeforeDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 2),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", "true"),
				),
			},
			{
				Config: testAccGroupConfig_destroyWhenProtectedFromScaleInAfterDestroy(rName),
				// Reaching this step is good enough, as it indicates the ASG was destroyed successfully.
			},
		},
	})
}

func TestAccAutoScalingGroup_mixedInstancesPolicy(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", "2"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicy_capacityRebalance(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyCapacityRebalance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", "true"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_requirements.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", "2"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_onDemandAllocationStrategy(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandAllocationStrategy(rName, "prioritized"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_allocation_strategy", "prioritized"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_onDemandBaseCapacity(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "1"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "2"),
				),
			},
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "0"),
				),
			},
		},
	})
}

// Test to verify fix for behavior in GH-ISSUE 7368
func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_updateToZeroOnDemandBaseCapacity(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "1"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "0"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_onDemandPercentageAboveBaseCapacity(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandPercentageAboveBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity", "1"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandPercentageAboveBaseCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity", "2"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_spotAllocationStrategy(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotAllocationStrategy(rName, "lowest-price"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_allocation_strategy", "lowest-price"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_spotInstancePools(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotInstancePools(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools", "2"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotInstancePools(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools", "3"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_spotMaxPrice(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotMaxPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", "0.50"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotMaxPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", "0.51"),
				),
			},
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotMaxPrice(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", ""),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateLaunchTemplateSpecification_launchTemplateName(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationLaunchTemplateName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateLaunchTemplateSpecification_version(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationVersion(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "1"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationVersion(rName, "$Latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Latest"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceType(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceType(rName, "t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceType(rName, "t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.medium"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceTypeWithLaunchTemplateSpecification(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceTypeLaunchTemplateSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckNoResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.launch_template_specification.#"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t4g.small"),
					resource.TestCheckResourceAttrPair(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.launch_template_specification.0.launch_template_id", "aws_launch_template.test-arm", "id"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_weightedCapacity(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideWeightedCapacity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", "4"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_weightedCapacity_withELB(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideWeightedCapacityELB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", "2"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_memoryMiBAndVCPUCount(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.0.min", "500"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.0.min", "1"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_mib {
                       min = 1000
                       max = 10000
                     }
                     vcpu_count {
                       min = 2
                       max = 12
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.0.min", "1000"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.0.max", "10000"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.0.min", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.0.max", "12"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorCount(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_count {
                       min = 2
                     }
                     memory_mib {
                      min = 500
                     }
                     vcpu_count {
                      min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.0.min", "2"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_count {
                       min = 1
                       max = 3
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.0.max", "3"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_count {
                       max = 0
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.0.max", "0"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorManufacturers(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_manufacturers = ["amazon-web-services"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amazon-web-services"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_manufacturers = ["amazon-web-services", "amd", "nvidia", "xilinx"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amazon-web-services"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "nvidia"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "xilinx"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorNames(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_names = ["a100"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "a100"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_names = ["a100", "v100", "k80", "t4", "m60", "radeon-pro-v520", "vu9p"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "a100"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "v100"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "k80"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "t4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "m60"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "radeon-pro-v520"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "vu9p"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorTotalMemoryMiB(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_total_memory_mib {
                       min = 32
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.min", "32"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_total_memory_mib {
                       max = 12000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.max", "12000"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_total_memory_mib {
                       min = 32
                       max = 12000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.min", "32"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.max", "12000"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorTypes(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_types = ["fpga"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.*", "fpga"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_types = ["fpga", "gpu", "inference"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.*", "fpga"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.*", "gpu"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.*", "inference"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_bareMetal(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`bare_metal = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.bare_metal", "excluded"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`bare_metal = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.bare_metal", "included"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`bare_metal = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.bare_metal", "required"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_baselineEBSBandwidthMbps(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`baseline_ebs_bandwidth_mbps {
                       min = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", "10"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`baseline_ebs_bandwidth_mbps {
                       max = 20000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.max", "20000"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`baseline_ebs_bandwidth_mbps {
                       min = 10
                       max = 20000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", "10"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.max", "20000"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_burstablePerformance(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`burstable_performance = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.burstable_performance", "excluded"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`burstable_performance = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.burstable_performance", "included"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`burstable_performance = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.burstable_performance", "required"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_cpuManufacturers(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`cpu_manufacturers = ["amazon-web-services"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amazon-web-services"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`cpu_manufacturers = ["amazon-web-services", "amd", "intel"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amazon-web-services"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.*", "intel"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_excludedInstanceTypes(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`excluded_instance_types = ["t2.nano"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.*", "t2.nano"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`excluded_instance_types = ["t2.nano", "t3*", "t4g.*"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.*", "t2.nano"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.*", "t3*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.*", "t4g.*"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_instanceGenerations(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`instance_generations = ["current"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.*", "current"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`instance_generations = ["current", "previous"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.*", "current"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.*", "previous"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_localStorage(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage", "excluded"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage", "included"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage", "required"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_localStorageTypes(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage_types = ["hdd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.*", "hdd"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage_types = ["hdd", "ssd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.*", "hdd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.*", "ssd"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_memoryGiBPerVCPU(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_gib_per_vcpu {
                       min = 0.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.min", "0.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_gib_per_vcpu {
                       max = 9.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.max", "9.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_gib_per_vcpu {
                       min = 0.5
                       max = 9.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.min", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.max", "9.5"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_networkInterfaceCount(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_interface_count {
                       min = 1
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.0.min", "1"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_interface_count {
                       max = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.0.max", "10"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_interface_count {
                       min = 1
                       max = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.0.max", "10"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_onDemandMaxPricePercentageOverLowestPrice(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`on_demand_max_price_percentage_over_lowest_price = 50
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.on_demand_max_price_percentage_over_lowest_price", "50"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_requireHibernateSupport(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`require_hibernate_support = false
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.require_hibernate_support", "false"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`require_hibernate_support = true
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.require_hibernate_support", "true"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_spotMaxPricePercentageOverLowestPrice(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`spot_max_price_percentage_over_lowest_price = 75
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.spot_max_price_percentage_over_lowest_price", "75"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_totalLocalStorageGB(t *testing.T) {
	var group autoscaling.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`total_local_storage_gb {
                       min = 0.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.0.min", "0.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`total_local_storage_gb {
                       max = 20.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.0.max", "20.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`total_local_storage_gb {
                       min = 0.5
                       max = 20.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.0.min", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.0.max", "20.5"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func testAccCheckGroupExists(n string, v *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Auto Scaling Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		output, err := tfautoscaling.FindGroupByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_group" {
			continue
		}

		_, err := tfautoscaling.FindGroupByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Auto Scaling Group %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckGroupHealthyInstanceCount(v *autoscaling.Group, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		count := 0

		for _, v := range v.Instances {
			if aws.StringValue(v.HealthStatus) == tfautoscaling.InstanceHealthStatusHealthy {
				count++
			}
		}

		if count < expected {
			return fmt.Errorf("Expected at least %d healthy instances, got %d", expected, count)
		}

		return nil
	}
}

func testAccCheckInstanceRefreshCount(v *autoscaling.Group, expected int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		output, err := tfautoscaling.FindInstanceRefreshes(conn, &autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: v.AutoScalingGroupName,
		})

		if err != nil {
			return err
		}

		if got := len(output); got != expected {
			return fmt.Errorf("Expected %d Instance Refreshes, got %d", expected, got)
		}

		return nil
	}
}

func testAccCheckInstanceRefreshStatus(v *autoscaling.Group, index int, expected ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		output, err := tfautoscaling.FindInstanceRefreshes(conn, &autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: v.AutoScalingGroupName,
		})

		if err != nil {
			return err
		}

		if got := len(output); got < index {
			return fmt.Errorf("Expected at least %d Instance Refreshes, got %d", index+1, got)
		}

		status := aws.StringValue(output[index].Status)

		for _, v := range expected {
			if status == v {
				return nil
			}
		}

		return fmt.Errorf("Expected Instance Refresh at index %d to be in %q, got %q", index, expected, status)
	}
}

func testAccCheckLBTargetGroupExists(n string, v *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No ELBv2 Target Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

		output, err := tfelbv2.FindTargetGroupByARN(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading ELBv2 Target Group (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("ELBv2 Target Group (%s) not found", rs.Primary.ID)
		}

		*v = *output

		return nil
	}
}

// testAccCheckALBTargetGroupHealthy checks an *elbv2.TargetGroup to make
// sure that all instances in it are healthy.
func testAccCheckALBTargetGroupHealthy(v *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

		output, err := conn.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
			TargetGroupArn: v.TargetGroupArn,
		})

		if err != nil {
			return err
		}

		for _, v := range output.TargetHealthDescriptions {
			if v.TargetHealth == nil || aws.StringValue(v.TargetHealth.State) != elbv2.TargetHealthStateEnumHealthy {
				return errors.New("Not all instances in target group are healthy yet, but should be")
			}
		}

		return nil
	}
}

func testAccGroupLaunchConfigurationBaseConfig(rName, instanceType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = %[2]q
}
`, rName, instanceType))
}

func testAccGroupLaunchTemplateBaseConfig(rName, instanceType string) string {
	// Include a Launch Configuration so that we can test swapping between Launch Template and Launch Configuration and vice-versa.
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, instanceType), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = %[2]q
}
`, rName, instanceType))
}

func testAccGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
}
`, rName))
}

func testAccGroupConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.test.name
}
`)
}

func testAccGroupConfig_namePrefix(rName, namePrefix string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name_prefix          = %[1]q
  launch_configuration = aws_launch_configuration.test.name
}
`, namePrefix))
}

func testAccGroupConfig_tags1(rName, tagKey1, tagValue1 string, tagPropagateAtLaunch1 bool) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = %[2]q
    value               = %[3]q
    propagate_at_launch = %[4]t
  }
}
`, rName, tagKey1, tagValue1, tagPropagateAtLaunch1))
}

func testAccGroupConfig_tags2(rName, tagKey1, tagValue1 string, tagPropagateAtLaunch1 bool, tagKey2, tagValue2 string, tagPropagateAtLaunch2 bool) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = %[2]q
    value               = %[3]q
    propagate_at_launch = %[4]t
  }

  tag {
    key                 = %[5]q
    value               = %[6]q
    propagate_at_launch = %[7]t
  }
}
`, rName, tagKey1, tagValue1, tagPropagateAtLaunch1, tagKey2, tagValue2, tagPropagateAtLaunch2))
}

func testAccGroupConfig_deprecatedTags1(rName, tagKey1, tagValue1 string, tagPropagateAtLaunch1 bool) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  tags = [{
    "key"                 = %[2]q
    "value"               = %[3]q
    "propagate_at_launch" = %[4]t
  }]
}
`, rName, tagKey1, tagValue1, tagPropagateAtLaunch1))
}

func testAccGroupConfig_simple(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_simpleUpdated(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_launch_configuration" "test2" {
  name          = "%[1]s-2"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = %[1]q
  max_size                  = 6
  min_size                  = 3
  health_check_grace_period = 400
  health_check_type         = "ELB"
  desired_capacity          = 4
  force_delete              = true
  termination_policies      = ["ClosestToNextInstanceHour"]
  protect_from_scale_in     = true

  launch_configuration = aws_launch_configuration.test2.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_terminationPoliciesExplicitDefault(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
  termination_policies = ["Default"]
}
`, rName))
}

func testAccGroupConfig_terminationPoliciesUpdated(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
  termination_policies = ["OldestInstance"]
}
`, rName))
}

func testAccGroupConfig_az(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  depends_on = [aws_subnet.test[0]]
}
`, rName))
}

func testAccGroupConfig_vpcZoneIdentifier(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
}
`, rName))
}

func testAccGroupELBBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_elb" "test" {
  name            = %[1]q
  subnets         = aws_subnet.test[*].id
  security_groups = [aws_security_group.test.id]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    target              = "HTTP:80/"
    interval            = 5
    timeout             = 2
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_launch_configuration" "test" {
  name            = %[1]q
  image_id        = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type   = "t2.micro"
  security_groups = [aws_security_group.test.id]

  # Need the instance to listen on port 80 at boot
  user_data = <<EOF
#!/bin/bash
echo "Terraform aws_autoscaling_group testing" > index.html
nohup python -m SimpleHTTPServer 80 &
EOF
}
`, rName))
}

func testAccGroupConfig_loadBalancer(rName string) string {
	return acctest.ConfigCompose(testAccGroupELBBaseConfig(rName), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 2
  min_size             = 2
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true
  load_balancers            = [aws_elb.test.name]

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_target2(rName string) string {
	return acctest.ConfigCompose(testAccGroupELBBaseConfig(rName), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 2
  min_size             = 2
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true
  target_group_arns         = [aws_lb_target_group.test.arn]

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_placement(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "c3.large"), fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = %[1]q
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 1
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  placement_group           = aws_placement_group.test.name
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_enabledMetrics(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  enabled_metrics = [
    "GroupTotalInstances",
    "GroupPendingInstances",
    "GroupTerminatingInstances",
    "GroupDesiredCapacity",
    "GroupMaxSize"
  ]
}
`, rName))
}

func testAccGroupConfig_allMetricsEnabled(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  enabled_metrics = [
    "GroupTotalInstances",
    "GroupPendingInstances",
    "GroupTerminatingInstances",
    "GroupDesiredCapacity",
    "GroupInServiceInstances",
    "GroupMinSize",
    "GroupMaxSize",
    "GroupPendingCapacity",
    "GroupInServiceCapacity",
    "GroupStandbyCapacity",
    "GroupTotalCapacity",
    "GroupTerminatingCapacity",
    "GroupStandbyInstances"
  ]
  metrics_granularity = "1Minute"
}
`, rName))
}

func testAccGroupConfig_suspendedProcesses(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration = aws_launch_configuration.test.name

  suspended_processes = ["AlarmNotification", "ScheduledActions"]

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_suspendedProcessesUpdated(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration = aws_launch_configuration.test.name

  suspended_processes = ["AZRebalance", "ScheduledActions"]

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_serviceLinkedRoleARN(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
data "aws_iam_role" "test" {
  name = "AWSServiceRoleForAutoScaling"
}

resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  service_linked_role_arn = data.aws_iam_role.test.arn
}
`, rName))
}

func testAccGroupConfig_maxInstanceLifetime(rName string, maxInstanceLifetime int) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  max_instance_lifetime = %[2]d
}
`, rName, maxInstanceLifetime))
}

func testAccGroupConfig_initialLifecycleHook(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration = aws_launch_configuration.test.name

  initial_lifecycle_hook {
    name                 = "launching"
    default_result       = "CONTINUE"
    heartbeat_timeout    = 30 # minimum value
    lifecycle_transition = "autoscaling:EC2_INSTANCE_LAUNCHING"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_launchTemplate(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}
`, rName))
}

func testAccGroupConfig_launchTemplateName(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    name = aws_launch_template.test.name
  }
}
`, rName))
}

func testAccGroupConfig_launchTemplateLatestVersion(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }
}
`, rName))
}

func testAccGroupConfig_largeDesiredCapacity(rName string, n int) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = %[2]d
  min_size             = %[2]d
  desired_capacity     = %[2]d
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName, n))
}

func testAccGroupConfig_instanceRefreshBasic(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshMinHealthyPercentage(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      min_healthy_percentage = 0
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshSkipMatching(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      min_healthy_percentage = 0
      skip_matching          = true
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshFull(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      instance_warmup        = 10
      min_healthy_percentage = 50
      checkpoint_delay       = 25
      checkpoint_percentages = [1, 20, 25, 50, 100]
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshDisabled(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshStart(rName, launchConfigurationNamePrefix string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}

resource "aws_launch_configuration" "test" {
  name_prefix   = %[2]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.nano"

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, launchConfigurationNamePrefix))
}

func testAccGroupConfig_instanceRefreshTriggers(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
    triggers = ["tags"]
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }

  tag {
    key                 = "Key"
    value               = "Value"
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_loadBalancers(rName string, elbCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_elb" "test" {
  count = %[2]d

  # "name" cannot be longer than 32 characters.
  name    = format("%%s-%%s", substr(%[1]q, 0, 28), count.index)
  subnets = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_autoscaling_group" "test" {
  name                = %[1]q
  force_delete        = true
  max_size            = 0
  min_size            = 0
  load_balancers      = length(aws_elb.test) > 0 ? aws_elb.test[*].name : []
  vpc_zone_identifier = aws_subnet.test[*].id

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, elbCount))
}

func testAccGroupConfig_target(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"

  enable_monitoring = false
}

resource "aws_lb_target_group" "test" {
  count = %[2]d

  name     = format("%%s-%%s", substr(%[1]q, 0, 28), count.index)
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  target_group_arns = length(aws_lb_target_group.test) > 0 ? aws_lb_target_group.test[*].arn : []
}
`, rName, targetGroupCount))
}

func testAccGroupConfig_targetELBCapacity(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name    = %[1]q
  subnets = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path              = "/"
    healthy_threshold = "2"
    timeout           = "2"
    interval          = "5"
    matcher           = "200"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count = length(aws_subnet.test[*])

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"

  associate_public_ip_address = "true"

  # Need the instance to listen on port 80 at boot
  user_data = <<EOF
#!/bin/bash
yum -y install httpd
echo "hello world" > /var/www/html/index.html
chkconfig httpd on
service httpd start
EOF
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier = aws_subnet.test[*].id

  target_group_arns = [aws_lb_target_group.test.arn]

  name                      = %[1]q
  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 2
  wait_for_elb_capacity     = 2
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_warmPoolEmpty(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  warm_pool {}

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_warmPoolFull(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  warm_pool {
    pool_state                  = "Stopped"
    min_size                    = 0
    max_group_prepared_capacity = 2
    instance_reuse_policy {
      reuse_on_scale_in = true
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_warmPoolNone(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_partition(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "m5.large"

  placement {
    tenancy    = "default"
    group_name = aws_placement_group.test.id
  }
}

resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }
}
`, rName))
}

func testAccGroupConfig_destroyWhenProtectedFromScaleInBeforeDestroy(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  name                  = %[1]q
  max_size              = 2
  min_size              = 2
  desired_capacity      = 2
  protect_from_scale_in = true
  launch_configuration  = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_destroyWhenProtectedFromScaleInAfterDestroy(rName string) string {
	return testAccGroupLaunchConfigurationBaseConfig(rName, "t3.micro")
}

func testAccGroupConfig_mixedInstancesPolicy(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "1"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyCapacityRebalance(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q
  capacity_rebalance = true

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "1"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandAllocationStrategy(rName, onDemandAllocationStrategy string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      on_demand_allocation_strategy = %[2]q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandAllocationStrategy))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName string, onDemandBaseCapacity int) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 2
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity = %[2]d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandBaseCapacity))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandPercentageAboveBaseCapacity(rName string, onDemandPercentageAboveBaseCapacity int) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      on_demand_percentage_above_base_capacity = %[2]d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandPercentageAboveBaseCapacity))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotAllocationStrategy(rName, spotAllocationStrategy string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      spot_allocation_strategy = %[2]q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotAllocationStrategy))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotInstancePools(rName string, spotInstancePools int) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      spot_instance_pools = %[2]d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotInstancePools))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotMaxPrice(rName, spotMaxPrice string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      spot_max_price = %[2]q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotMaxPrice))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationLaunchTemplateName(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_name = aws_launch_template.test.name
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationVersion(rName, version string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
        version            = %[2]q
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, version))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceType(rName, instanceType string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = %[2]q
      }
    }
  }
}
`, rName, instanceType))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceTypeLaunchTemplateSpecification(rName string) string {
	return acctest.ConfigCompose(
		testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"),
		acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI(),
		fmt.Sprintf(`
resource "aws_launch_template" "test-arm" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.micro"
  name          = "%[1]s-arm"
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t4g.small"

        launch_template_specification {
          launch_template_id = aws_launch_template.test-arm.id
        }
      }
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideWeightedCapacity(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 4
  max_size           = 6
  min_size           = 2
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "4"
      }
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideWeightedCapacityELB(rName string) string {
	return acctest.ConfigCompose(testAccGroupELBBaseConfig(rName), fmt.Sprintf(`
locals {
  user_data = <<EOF
#!/bin/bash
echo "Terraform aws_autoscaling_group Testing" > index.html
nohup python -m SimpleHTTPServer 80 &
EOF
}

resource "aws_launch_template" "test" {
  image_id               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type          = "t3.micro"
  name                   = %[1]q
  user_data              = base64encode(local.user_data)
  vpc_security_group_ids = [aws_security_group.test.id]
}

resource "aws_autoscaling_group" "test" {
  desired_capacity      = 2
  wait_for_elb_capacity = 2
  max_size              = 2
  min_size              = 2
  name                  = %[1]q
  load_balancers        = [aws_elb.test.name]
  vpc_zone_identifier   = aws_subnet.test[*].id

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t3.micro"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName string, instanceRequirements string) string {
	return acctest.ConfigCompose(testAccGroupLaunchTemplateBaseConfig(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_requirements {
          %[2]s
        }
      }
    }

    instances_distribution {
      on_demand_percentage_above_base_capacity = 50
      spot_allocation_strategy                 = "capacity-optimized"
    }
  }
}
`, rName, instanceRequirements))
}
