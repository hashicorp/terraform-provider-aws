package autoscaling_test

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
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
				Config: testAccGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf(`autoScalingGroup:.+:autoScalingGroupName/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", "false"),
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
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupConfig(rName),
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
				Config: testAccGroupNameGeneratedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupNamePrefixConfig(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupConfigTags1(rName, "key1", "value1", true),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfigTags2(rName, "key1", "value1updated", true, "key2", "value2", false),
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
				Config: testAccGroupConfigTags1(rName, "key2", "value2", true),
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
				Config: testAccGroupConfigDeprecatedTags1(rName, "key1", "value1", true),
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
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupSimpleConfig(rName),
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
					resource.TestCheckNoResourceAttr(resourceName, "tag.#"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupSimpleUpdatedConfig(rName),
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
					resource.TestCheckNoResourceAttr(resourceName, "tag.#"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
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
				Config: testAccGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "0"),
				),
			},
			{
				Config: testAccGroupTerminationPoliciesUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "OldestInstance"),
				),
			},
			{
				Config: testAccGroupTerminationPoliciesExplicitDefaultConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "Default"),
				),
			},
			{
				Config: testAccGroupConfig(rName),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithAZConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("aws_autoscaling_group.bar", "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "vpc_zone_identifier.#", "0"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
			{
				Config: testAccGroupWithVPCIdentConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckGroupAttributesVPCZoneIdentifier(&group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("aws_autoscaling_group.bar", "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "vpc_zone_identifier.#", "1"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withLoadBalancer(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithLoadBalancerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckGroupAttributesLoadBalancer(&group),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
		},
	})
}

func TestAccAutoScalingGroup_WithLoadBalancer_toTargetGroup(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithLoadBalancerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupWithTargetGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupWithLoadBalancerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_withPlacementGroup(t *testing.T) {
	var group autoscaling.Group

	randName := fmt.Sprintf("tf-test-%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_withPlacementGroup(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "placement_group", randName),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
		},
	})
}

func TestAccAutoScalingGroup_enablingMetrics(t *testing.T) {
	var group autoscaling.Group
	randName := fmt.Sprintf("terraform-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupSimpleConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckNoResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
			{
				Config: testAccMetricsCollectionConfig_updatingMetricsCollected(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics.#", "5"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_suspendingProcesses(t *testing.T) {
	var group autoscaling.Group
	randName := fmt.Sprintf("terraform-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupSimpleConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "suspended_processes.#", "0"),
				),
			},
			{
				Config: testAccGroupWithSuspendedProcessesConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "suspended_processes.#", "2"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
			{
				Config: testAccGroupWithSuspendedProcessesUpdatedConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "suspended_processes.#", "2"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withMetrics(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricsCollectionConfig_allMetricsCollected(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics.#", "13"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
			{
				Config: testAccMetricsCollectionConfig_updatingMetricsCollected(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics.#", "5"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_serviceLinkedRoleARN(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_withServiceLinkedRoleARN(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "service_linked_role_arn"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
		},
	})
}

func TestAccAutoScalingGroup_maxInstanceLifetime(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_withMaxInstanceLifetime(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "max_instance_lifetime", "864000"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
			{
				Config: testAccGroupConfig_withMaxInstanceLifetime_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "max_instance_lifetime", "604800"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_ALB_targetGroups(t *testing.T) {
	var group autoscaling.Group
	var tg elbv2.TargetGroup
	var tg2 elbv2.TargetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func(targets []*elbv2.TargetGroup) resource.TestCheckFunc {
		return func(*terraform.State) error {
			var ts []string
			var gs []string
			for _, t := range targets {
				ts = append(ts, *t.TargetGroupArn)
			}

			for _, s := range group.TargetGroupARNs {
				gs = append(gs, *s)
			}

			sort.Strings(ts)
			sort.Strings(gs)

			if !reflect.DeepEqual(ts, gs) {
				return fmt.Errorf("Error: target group match not found!\nASG Target groups: %#v\nTarget Group: %#v", ts, gs)
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_ALB_TargetGroup_pre(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckLBTargetGroupExists("aws_lb_target_group.test", &tg),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "target_group_arns.#", "0"),
				),
			},

			{
				Config: testAccGroupConfig_ALB_TargetGroup_post_duo(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckLBTargetGroupExists("aws_lb_target_group.test", &tg),
					testAccCheckLBTargetGroupExists("aws_lb_target_group.test_more", &tg2),
					testCheck([]*elbv2.TargetGroup{&tg, &tg2}),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "target_group_arns.#", "2"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
			{
				Config: testAccGroupConfig_ALB_TargetGroup_post(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckLBTargetGroupExists("aws_lb_target_group.test", &tg),
					testCheck([]*elbv2.TargetGroup{&tg}),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "target_group_arns.#", "1"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/256
func TestAccAutoScalingGroup_targetGroupARNs(t *testing.T) {
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
				Config: testAccGroupConfig_TargetGroupARNs(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "11"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_TargetGroupARNs(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_TargetGroupARNs(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "11"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_initialLifecycleHook(t *testing.T) {
	var group autoscaling.Group

	randName := fmt.Sprintf("terraform-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithHookConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckGroupHealthyCapacity(&group, 2),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "initial_lifecycle_hook.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("aws_autoscaling_group.bar", "initial_lifecycle_hook.*", map[string]string{
						"default_result": "CONTINUE",
						"name":           "launching",
					}),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
		},
	})
}

func TestAccAutoScalingGroup_ALBTargetGroups_elbCapacity(t *testing.T) {
	var group autoscaling.Group
	var tg elbv2.TargetGroup

	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_ALB_TargetGroup_ELBCapacity(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckLBTargetGroupExists("aws_lb_target_group.test", &tg),
					testAccCheckALBTargetGroupHealthy(&tg),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_basic(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_InstanceRefresh_Basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"wait_for_capacity_timeout",
					"instance_refresh",
				},
			},
			{
				Config: testAccGroupConfig_InstanceRefresh_MinHealthyPercentage(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_InstanceRefresh_SkipMatching(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", "true"),
				),
			},
			{
				Config: testAccGroupConfig_InstanceRefresh_Full(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", "10"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", "25"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.0", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.1", "20"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.2", "25"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.3", "50"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.4", "100"),
				),
			},
			{
				Config: testAccGroupConfig_InstanceRefresh_Disabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckNoResourceAttr(resourceName, "instance_refresh.#"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_start(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	launchConfigurationName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_InstanceRefresh_Start("one"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationName, "name"),
					testAccCheckInstanceRefreshCount(&group, 0),
				),
			},
			{
				Config: testAccGroupConfig_InstanceRefresh_Start("two"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationName, "name"),
					testAccCheckInstanceRefreshCount(&group, 1),
					testAccCheckInstanceRefreshStatus(&group, 0, autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress),
				),
			},
			{
				Config: testAccGroupConfig_InstanceRefresh_Start("three"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationName, "name"),
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
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_InstanceRefresh_Basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_InstanceRefresh_Triggers(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_refresh.0.triggers.*", "tags"),
					testAccCheckInstanceRefreshCount(&group, 1),
					testAccCheckInstanceRefreshStatus(&group, 0, autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_warmPool(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_WarmPool_Empty(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"warm_pool",
					"force_delete",
					"wait_for_capacity_timeout",
				},
			},
			{
				Config: testAccGroupConfig_WarmPool_Full(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.pool_state", "Stopped"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.min_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.max_group_prepared_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.0.reuse_on_scale_in", "true"),
				),
			},
			{
				Config: testAccGroupConfig_WarmPool_Remove(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckNoResourceAttr(resourceName, "warm_pool.#"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_classicVPCZoneIdentifier(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_classicVPCZoneIdentifier(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.test", &group),
					resource.TestCheckResourceAttr("aws_autoscaling_group.test", "vpc_zone_identifier.#", "0"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.test",
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
			},
		},
	})
}

func TestAccAutoScalingGroup_launchTemplate(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_withLaunchTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.id"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
		},
	})
}

func TestAccAutoScalingGroup_LaunchTemplate_update(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_withLaunchTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.name"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
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
			},
			{
				Config: testAccGroupConfig_withLaunchTemplate_toLaunchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_configuration"),
					resource.TestCheckNoResourceAttr(
						"aws_autoscaling_group.bar", "launch_template"),
				),
			},

			{
				Config: testAccGroupConfig_withLaunchTemplate_toLaunchTemplateName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_configuration", ""),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.name", "foobar2"),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.id"),
				),
			},

			{
				Config: testAccGroupConfig_withLaunchTemplate_toLaunchTemplateVersion(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "$Latest"),
				),
			},

			{
				Config: testAccGroupConfig_withLaunchTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.name"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "1"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_LaunchTemplate_iamInstanceProfile(t *testing.T) {
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
				Config: testAccGroupConfig_LaunchTemplate_IAMInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
				),
			},
			{
				ResourceName:      resourceName,
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
				Config: testAccGroupConfig_LoadBalancers(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "11"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_LoadBalancers(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
				),
			},
			{
				Config: testAccGroupConfig_LoadBalancers(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "11"),
				),
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
				Config: testAccGroupConfig_MixedInstancesPolicy(rName),
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
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupConfig_MixedInstancesPolicy_CapacityRebalance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", "true"),
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
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandAllocationStrategy(rName, "prioritized"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_allocation_strategy", "prioritized"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "2"),
				),
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 0),
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
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "0"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandPercentageAboveBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandPercentageAboveBaseCapacity(rName, 2),
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
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotAllocationStrategy(rName, "lowest-price"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_allocation_strategy", "lowest-price"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotInstancePools(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools", "2"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotInstancePools(rName, 3),
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
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", "0.50"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", "0.51"),
				),
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(rName, ""),
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
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_LaunchTemplateName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_Version(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_Version(rName, "$Latest"),
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
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType(rName, "t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
				),
			},
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType(rName, "t3.medium"),
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
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType_With_LaunchTemplateSpecification(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckNoResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.launch_template_specification.#"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t4g.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.launch_template_specification.0.launch_template_id", "aws_launch_template.testarm", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
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
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_WeightedCapacity(rName),
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
			{
				ResourceName:      resourceName,
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
			},
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
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_WeightedCapacity_WithELB(rName),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_memoryMiBAndVCpuCount(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorCount(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorManufacturers(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorNames(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorTotalMemoryMiB(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorTypes(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_bareMetal(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_baselineEbsBandwidthMbps(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_burstablePerformance(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_cpuManufacturers(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_excludedInstanceTypes(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_instanceGenerations(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_localStorage(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_localStorageTypes(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_memoryGiBPerVCpu(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_networkInterfaceCount(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_onDemandMaxPricePercentageOverLowestPrice(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_requireHibernateSupport(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_spotMaxPricePercentageOverLowestPrice(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_totalLocalStorageGB(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
			{
				Config: testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
			{
				ResourceName:      resourceName,
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
			},
		},
	})
}

func TestAccAutoScalingGroup_launchTempPartitionNum(t *testing.T) {
	var group autoscaling.Group

	randName := fmt.Sprintf("terraform-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPartitionConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("aws_autoscaling_group.test", &group),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.test",
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
			},
		},
	})
}

func TestAccAutoScalingGroup_Destroy_whenProtectedFromScaleIn(t *testing.T) {
	var group autoscaling.Group
	rName := fmt.Sprintf("terraform-test-%s", sdkacctest.RandString(10))
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_DestroyWhenProtectedFromScaleIn_beforeDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					testAccCheckGroupHealthyCapacity(&group, 2),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", "true"),
				),
			},
			{
				Config: testAccGroupConfig_DestroyWhenProtectedFromScaleIn_afterDestroy(),
				// Reaching this step is good enough, as it indicates the ASG was destroyed successfully.
			},
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

func testAccCheckGroupAttributesLoadBalancer(group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(group.LoadBalancerNames) != 1 {
			return fmt.Errorf("Bad load_balancers: %v", group.LoadBalancerNames)
		}

		return nil
	}
}

func testAccCheckGroupHealthyCapacity(g *autoscaling.Group, exp int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		healthy := 0
		for _, i := range g.Instances {
			if i.HealthStatus == nil {
				continue
			}
			if strings.EqualFold(*i.HealthStatus, "Healthy") {
				healthy++
			}
		}
		if healthy < exp {
			return fmt.Errorf("Expected at least %d healthy, got %d.", exp, healthy)
		}
		return nil
	}
}

func testAccCheckGroupAttributesVPCZoneIdentifier(group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Grab Subnet Ids
		var subnets []string
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_subnet" {
				continue
			}
			subnets = append(subnets, rs.Primary.Attributes["id"])
		}

		if group.VPCZoneIdentifier == nil {
			return fmt.Errorf("Bad VPC Zone Identifier\nexpected: %s\ngot nil", subnets)
		}

		zones := strings.Split(*group.VPCZoneIdentifier, ",")

		remaining := len(zones)
		for _, z := range zones {
			for _, s := range subnets {
				if z == s {
					remaining--
				}
			}
		}

		if remaining != 0 {
			return fmt.Errorf("Bad VPC Zone Identifier match\nexpected: %s\ngot:%s", zones, subnets)
		}

		return nil
	}
}

// testAccCheckALBTargetGroupHealthy checks an *elbv2.TargetGroup to make
// sure that all instances in it are healthy.
func testAccCheckALBTargetGroupHealthy(res *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

		resp, err := conn.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
			TargetGroupArn: res.TargetGroupArn,
		})

		if err != nil {
			return err
		}

		for _, target := range resp.TargetHealthDescriptions {
			if target.TargetHealth == nil || target.TargetHealth.State == nil || *target.TargetHealth.State != "healthy" {
				return errors.New("Not all instances in target group are healthy yet, but should be")
			}
		}

		return nil
	}
}

func testAccGroupLaunchConfigurationBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}
`, rName))
}

func testAccGroupConfig(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
}
`, rName))
}

func testAccGroupNameGeneratedConfig(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.test.name
}
`)
}

func testAccGroupNamePrefixConfig(rName, namePrefix string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name_prefix          = %[1]q
  launch_configuration = aws_launch_configuration.test.name
}
`, namePrefix))
}

func testAccGroupConfigTags1(rName, tagKey1, tagValue1 string, tagPropagateAtLaunch1 bool) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
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

func testAccGroupConfigTags2(rName, tagKey1, tagValue1 string, tagPropagateAtLaunch1 bool, tagKey2, tagValue2 string, tagPropagateAtLaunch2 bool) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
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

func testAccGroupConfigDeprecatedTags1(rName, tagKey1, tagValue1 string, tagPropagateAtLaunch1 bool) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
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

func testAccGroupSimpleConfig(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
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
}
`, rName))
}

func testAccGroupSimpleUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
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
}
`, rName))
}

func testAccGroupTerminationPoliciesExplicitDefaultConfig(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
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

func testAccGroupTerminationPoliciesUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(testAccGroupLaunchConfigurationBaseConfig(rName), fmt.Sprintf(`
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

func testAccGroupWithLoadBalancerConfig() string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-lb"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_subnet" "foo" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.foo.id
  tags = {
    Name = "tf-acc-autoscaling-group-with-load-balancer"
  }
}

resource "aws_security_group" "foo" {
  vpc_id = aws_vpc.foo.id

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
}

resource "aws_elb" "bar" {
  subnets         = [aws_subnet.foo.id]
  security_groups = [aws_security_group.foo.id]

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

  depends_on = [aws_internet_gateway.gw]
}

resource "aws_launch_configuration" "foobar" {
  image_id        = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type   = "t2.micro"
  security_groups = [aws_security_group.foo.id]

  # Need the instance to listen on port 80 at boot
  user_data = <<EOF
#!/bin/bash
echo "Terraform aws_autoscaling_group Testing" > index.html
nohup python -m SimpleHTTPServer 80 &
EOF
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier       = [aws_subnet.foo.id]
  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true

  launch_configuration = aws_launch_configuration.foobar.name
  load_balancers       = [aws_elb.bar.name]
}
`)
}

func testAccGroupWithTargetGroupConfig() string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-lb"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_subnet" "foo" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.foo.id
  tags = {
    Name = "tf-acc-autoscaling-group-with-target-group"
  }
}

resource "aws_security_group" "foo" {
  vpc_id = aws_vpc.foo.id

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
}

resource "aws_lb_target_group" "foo" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.foo.id
}

resource "aws_elb" "bar" {
  subnets         = [aws_subnet.foo.id]
  security_groups = [aws_security_group.foo.id]

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

  depends_on = [aws_internet_gateway.gw]
}

resource "aws_launch_configuration" "foobar" {
  image_id        = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type   = "t2.micro"
  security_groups = [aws_security_group.foo.id]

  # Need the instance to listen on port 80 at boot
  user_data = <<EOF
#!/bin/bash
echo "Terraform aws_autoscaling_group Testing" > index.html
nohup python -m SimpleHTTPServer 80 &
EOF
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier       = [aws_subnet.foo.id]
  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true

  launch_configuration = aws_launch_configuration.foobar.name
  target_group_arns    = [aws_lb_target_group.foo.arn]
}
`)
}

func testAccGroupWithAZConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-az"
  }
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-autoscaling-group-with-az"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [
    data.aws_availability_zones.available.names[0]
  ]
  desired_capacity     = 0
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.foobar.name
}
`)
}

func testAccGroupWithVPCIdentConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-vpc-id"
  }
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-autoscaling-group-with-vpc-id"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
  ]
  desired_capacity     = 0
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.foobar.name
}
`)
}

func testAccGroupConfig_withPlacementGroup(name string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "c3.large"
}

resource "aws_placement_group" "test" {
  name     = "%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = "%s"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 1
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  placement_group           = aws_placement_group.test.name

  launch_configuration = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

func testAccGroupConfig_withServiceLinkedRoleARN() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_iam_role" "autoscaling_service_linked_role" {
  name = "AWSServiceRoleForAutoScaling"
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones      = [data.aws_availability_zones.available.names[0]]
  desired_capacity        = 0
  max_size                = 0
  min_size                = 0
  launch_configuration    = aws_launch_configuration.foobar.name
  service_linked_role_arn = data.aws_iam_role.autoscaling_service_linked_role.arn
}
`)
}

func testAccGroupConfig_withMaxInstanceLifetime() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  desired_capacity      = 0
  max_size              = 0
  min_size              = 0
  launch_configuration  = aws_launch_configuration.foobar.name
  max_instance_lifetime = "864000"
}
`)
}

func testAccGroupConfig_withMaxInstanceLifetime_update() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  desired_capacity      = 0
  max_size              = 0
  min_size              = 0
  launch_configuration  = aws_launch_configuration.foobar.name
  max_instance_lifetime = "604800"
}
`)
}

func testAccMetricsCollectionConfig_allMetricsCollected() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  max_size                  = 1
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "EC2"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration      = aws_launch_configuration.foobar.name
  enabled_metrics = ["GroupTotalInstances",
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
`)
}

func testAccMetricsCollectionConfig_updatingMetricsCollected() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  max_size                  = 1
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "EC2"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration      = aws_launch_configuration.foobar.name
  enabled_metrics = ["GroupTotalInstances",
    "GroupPendingInstances",
    "GroupTerminatingInstances",
    "GroupDesiredCapacity",
    "GroupMaxSize"
  ]
  metrics_granularity = "1Minute"
}
`)
}

func testAccGroupConfig_ALB_TargetGroup_pre(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-autoscaling-group-alb-target-group"
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.default.id
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = data.aws_ami.test_ami.id
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
    aws_subnet.alt.id,
  ]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name
}

resource "aws_security_group" "tf_test_self" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.default.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccGroupConfig_ALB_TargetGroup_post(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.default.id
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s-2"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = data.aws_ami.test_ami.id
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
    aws_subnet.alt.id,
  ]

  target_group_arns = [aws_lb_target_group.test.arn]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name
}

resource "aws_security_group" "tf_test_self" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.default.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccGroupConfig_ALB_TargetGroup_post_duo(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.default.id
}

resource "aws_lb_target_group" "test_more" {
  name     = format("%%s-%%s", substr(%[1]q, 0, 28), "2")
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.default.id
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s-2"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = data.aws_ami.test_ami.id
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
    aws_subnet.alt.id,
  ]

  target_group_arns = [
    aws_lb_target_group.test.arn,
    aws_lb_target_group.test_more.arn,
  ]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name
}

resource "aws_security_group" "tf_test_self" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.default.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccGroupConfig_TargetGroupARNs(rName string, tgCount int) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  count = %[2]d

  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_autoscaling_group" "test" {
  force_delete        = true
  max_size            = 0
  min_size            = 0
  target_group_arns   = length(aws_lb_target_group.test) > 0 ? aws_lb_target_group.test[*].arn : []
  vpc_zone_identifier = [aws_subnet.test.id]

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, tgCount)
}

func testAccGroupWithHookConfig(name string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = "%s"
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]

  launch_configuration = aws_launch_configuration.foobar.name

  initial_lifecycle_hook {
    name                 = "launching"
    default_result       = "CONTINUE"
    heartbeat_timeout    = 30 # minimum value
    lifecycle_transition = "autoscaling:EC2_INSTANCE_LAUNCHING"
  }
}
`, name)
}

func testAccGroupConfig_ALB_TargetGroup_ELBCapacity(rInt int) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = "true"
  enable_dns_support   = "true"

  tags = {
    Name = "terraform-testacc-autoscaling-group-alb-target-group-elb-capacity"
  }
}

resource "aws_lb" "test_lb" {
  subnets = [aws_subnet.main.id, aws_subnet.alt.id]

  tags = {
    Name = "testAccGroupConfig_ALB_TargetGroup_ELBCapacity"
  }
}

resource "aws_lb_listener" "test_listener" {
  load_balancer_arn = aws_lb.test_lb.arn
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "tf-alb-test-%d"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.default.id

  health_check {
    path              = "/"
    healthy_threshold = "2"
    timeout           = "2"
    interval          = "5"
    matcher           = "200"
  }

  tags = {
    Name = "testAccGroupConfig_ALB_TargetGroup_ELBCapacity"
  }
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-elb-capacity-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-elb-capacity-alt"
  }
}

resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.default.id
}

resource "aws_route_table" "route_table" {
  vpc_id = aws_vpc.default.id
}

resource "aws_route_table_association" "route_table_association_main" {
  subnet_id      = aws_subnet.main.id
  route_table_id = aws_route_table.route_table.id
}

resource "aws_route_table_association" "route_table_association_alt" {
  subnet_id      = aws_subnet.alt.id
  route_table_id = aws_route_table.route_table.id
}

resource "aws_route" "public_default_route" {
  route_table_id         = aws_route_table.route_table.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet_gateway.id
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id                    = data.aws_ami.test_ami.id
  instance_type               = "t2.micro"
  associate_public_ip_address = "true"

  user_data = <<EOS
#!/bin/bash
yum -y install httpd
echo "hello world" > /var/www/html/index.html
chkconfig httpd on
service httpd start
EOS
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
    aws_subnet.alt.id,
  ]

  target_group_arns = [aws_lb_target_group.test.arn]

  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 2
  wait_for_elb_capacity     = 2
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name
}
`, rInt)
}

func testAccGroupWithSuspendedProcessesConfig(name string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_placement_group" "test" {
  name     = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = "%s"
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]

  launch_configuration = aws_launch_configuration.foobar.name

  suspended_processes = ["AlarmNotification", "ScheduledActions"]

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

func testAccGroupWithSuspendedProcessesUpdatedConfig(name string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_placement_group" "test" {
  name     = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = "%s"
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]

  launch_configuration = aws_launch_configuration.foobar.name

  suspended_processes = ["AZRebalance", "ScheduledActions"]

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

func testAccGroupConfig_classicVPCZoneIdentifier() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
resource "aws_autoscaling_group" "test" {
  min_size = 0
  max_size = 0

  availability_zones   = [data.aws_availability_zones.available.names[0]]
  launch_configuration = aws_launch_configuration.test.name
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t1.micro"
}
`)
}

func testAccGroupConfig_withLaunchTemplate() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  launch_template {
    id      = aws_launch_template.foobar.id
    version = aws_launch_template.foobar.default_version
  }
}
`)
}

func testAccGroupConfig_withLaunchTemplate_toLaunchConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  desired_capacity     = 0
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.test.name
}
`)
}

func testAccGroupConfig_withLaunchTemplate_toLaunchTemplateName() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_template" "foobar2" {
  name          = "foobar2"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  launch_template {
    name = "foobar2"
  }
}
`)
}

func testAccGroupConfig_withLaunchTemplate_toLaunchTemplateVersion() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_template" "foobar2" {
  name          = "foobar2"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  launch_template {
    id      = aws_launch_template.foobar.id
    version = "$Latest"
  }
}
`)
}

func testAccGroupConfig_LaunchTemplate_IAMInstanceProfile(rName string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
  name               = %q
}

resource "aws_iam_instance_profile" "test" {
  name = %q
  role = aws_iam_role.test.name
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t2.micro"
  name          = %q

  iam_instance_profile {
    name = aws_iam_instance_profile.test.id
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, rName, rName, rName)
}

func testAccGroupConfig_LoadBalancers(rName string, elbCount int) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_elb" "test" {
  count      = %[2]d
  depends_on = [aws_internet_gateway.test]

  subnets = [aws_subnet.test.id]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_autoscaling_group" "test" {
  force_delete        = true
  max_size            = 0
  min_size            = 0
  load_balancers      = length(aws_elb.test) > 0 ? aws_elb.test[*].name : []
  vpc_zone_identifier = [aws_subnet.test.id]

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, elbCount)
}

func testAccGroupConfig_MixedInstancesPolicy_Base(rName string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %q
}
`, rName)
}

func testAccGroupConfig_MixedInstancesPolicy_Arm_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "testarm" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-arm64-gp2"]
  }
}

resource "aws_launch_template" "testarm" {
  image_id      = data.aws_ami.testarm.id
  instance_type = "t4g.micro"
  name          = %q
}
`, rName)
}

func testAccGroupConfig_MixedInstancesPolicy(rName string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

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
`, rName)
}

func testAccGroupConfig_MixedInstancesPolicy_CapacityRebalance(rName string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  capacity_rebalance = true
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

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
`, rName)
}

func testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandAllocationStrategy(rName, onDemandAllocationStrategy string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      on_demand_allocation_strategy = %q
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
`, rName, onDemandAllocationStrategy)
}

func testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName string, onDemandBaseCapacity int) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 2
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity = %d
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
`, rName, onDemandBaseCapacity)
}

func testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandPercentageAboveBaseCapacity(rName string, onDemandPercentageAboveBaseCapacity int) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      on_demand_percentage_above_base_capacity = %d
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
`, rName, onDemandPercentageAboveBaseCapacity)
}

func testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotAllocationStrategy(rName, spotAllocationStrategy string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      spot_allocation_strategy = %q
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
`, rName, spotAllocationStrategy)
}

func testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotInstancePools(rName string, spotInstancePools int) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      spot_instance_pools = %d
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
`, rName, spotInstancePools)
}

func testAccGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(rName, spotMaxPrice string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      spot_max_price = %q
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
`, rName, spotMaxPrice)
}

func testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_LaunchTemplateName(rName string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

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
`, rName)
}

func testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_Version(rName, version string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
        version            = %q
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
`, rName, version)
}

func testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType(rName, instanceType string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = %q
      }
    }
  }
}
`, rName, instanceType)
}

func testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType_With_LaunchTemplateSpecification(rName, rName2 string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		testAccGroupConfig_MixedInstancesPolicy_Arm_Base(rName2) +
		fmt.Sprintf(`

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t4g.micro"
        launch_template_specification {
          launch_template_id = aws_launch_template.testarm.id
        }
      }
    }
  }
}
`, rName)
}

func testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_WeightedCapacity(rName string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 4
  max_size           = 6
  min_size           = 2
  name               = %q

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
}
`, rName)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_WeightedCapacity_WithELB(rName string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = %[1]q
  }
}
resource "aws_subnet" "foo" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "foo" {
  vpc_id = aws_vpc.foo.id

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
}

resource "aws_lb_target_group" "foo" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.foo.id
}

resource "aws_elb" "bar" {
  subnets         = [aws_subnet.foo.id]
  security_groups = [aws_security_group.foo.id]
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
  depends_on = [aws_internet_gateway.gw]
}

locals {
  user_data = <<EOF
  #!/bin/bash
  echo "Terraform aws_autoscaling_group Testing" > index.html
  nohup python -m SimpleHTTPServer 80 &
  EOF
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]
  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id               = data.aws_ami.test.id
  instance_type          = "t3.micro"
  name                   = %[1]q
  user_data              = base64encode(local.user_data)
  vpc_security_group_ids = [aws_security_group.foo.id]
}

resource "aws_autoscaling_group" "test" {
  desired_capacity      = 2
  wait_for_elb_capacity = 2
  max_size              = 2
  min_size              = 2
  name                  = %[1]q
  load_balancers        = [aws_elb.bar.name]
  vpc_zone_identifier   = aws_subnet.foo[*].id
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
}
`, rName)
}

func testAccGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_instanceRequirements(rName string, instanceRequirements string) string {
	return testAccGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_requirements {
          %s
        }
      }
    }

    instances_distribution {
      on_demand_percentage_above_base_capacity = 50
      spot_allocation_strategy                 = "capacity-optimized"
    }
  }
}
`, rName, instanceRequirements)
}

func testAccGroupPartitionConfig(rName string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "this" {
  name          = %[1]q
  image_id      = data.aws_ami.test_ami.id
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
  name_prefix        = "test"
  availability_zones = [data.aws_availability_zones.available.names[0]]
  min_size           = 0
  max_size           = 0

  launch_template {
    id      = aws_launch_template.this.id
    version = "$Latest"
  }
}
`, rName)
}

func testAccGroupConfig_InstanceRefresh_Basic() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
  }
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccGroupConfig_InstanceRefresh_MinHealthyPercentage() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
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
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccGroupConfig_InstanceRefresh_SkipMatching() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
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
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccGroupConfig_InstanceRefresh_Full() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
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
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccGroupConfig_InstanceRefresh_Disabled() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccGroupConfig_InstanceRefresh_Start(launchConfigurationName string) string {
	return fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
  }
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  name_prefix   = %[1]q
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"

  lifecycle {
    create_before_destroy = true
  }
}
`, launchConfigurationName)
}

func testAccGroupConfig_InstanceRefresh_Triggers() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
    triggers = ["tags"]
  }

  tag {
    key                 = "Key"
    value               = "Value"
    propagate_at_launch = true
  }
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccGroupConfig_WarmPool_Base() string {
	return `
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccGroupConfig_WarmPool_Empty() string {
	return testAccGroupConfig_WarmPool_Base() + `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  warm_pool {}
}
`
}

func testAccGroupConfig_WarmPool_Full() string {
	return testAccGroupConfig_WarmPool_Base() + `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  warm_pool {
    pool_state                  = "Stopped"
    min_size                    = 0
    max_group_prepared_capacity = 2
    instance_reuse_policy {
      reuse_on_scale_in = true
    }
  }
}
`
}

func testAccGroupConfig_WarmPool_Remove() string {
	return testAccGroupConfig_WarmPool_Base() + `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name
}
`
}

func testAccGroupConfig_DestroyWhenProtectedFromScaleIn_beforeDestroy(name string) string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  name                  = %[1]q
  max_size              = 2
  min_size              = 2
  desired_capacity      = 2
  protect_from_scale_in = true
  launch_configuration  = aws_launch_configuration.test.name
}
`, name)
}

func testAccGroupConfig_DestroyWhenProtectedFromScaleIn_afterDestroy() string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() +
		`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
}
`
}

func testAccCheckInstanceRefreshCount(group *autoscaling.Group, expected int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		input := autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
		}
		resp, err := conn.DescribeInstanceRefreshes(&input)
		if err != nil {
			return fmt.Errorf("error describing Auto Scaling Group (%s) Instance Refreshes: %w", aws.StringValue(group.AutoScalingGroupName), err)
		}

		if len(resp.InstanceRefreshes) != expected {
			return fmt.Errorf("expected %d Instance Refreshes, got %d", expected, len(resp.InstanceRefreshes))
		}
		return nil
	}
}

func testAccCheckInstanceRefreshStatus(group *autoscaling.Group, offset int, expected ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		input := autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
		}
		resp, err := conn.DescribeInstanceRefreshes(&input)
		if err != nil {
			return fmt.Errorf("error describing Auto Scaling Group (%s) Instance Refreshes: %w", aws.StringValue(group.AutoScalingGroupName), err)
		}

		if len(resp.InstanceRefreshes) < offset {
			return fmt.Errorf("expected at least %d Instance Refreshes, got %d", offset+1, len(resp.InstanceRefreshes))
		}

		actual := aws.StringValue(resp.InstanceRefreshes[offset].Status)
		for _, s := range expected {
			if actual == s {
				return nil
			}
		}
		return fmt.Errorf("expected Instance Refresh at index %d to be in %q, got %q", offset, expected, actual)
	}
}

func TestCreateAutoScalingGroupInstanceRefreshInput(t *testing.T) {
	const asgName = "test-asg"
	testCases := []struct {
		name     string
		input    []interface{}
		expected *autoscaling.StartInstanceRefreshInput
	}{
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "nil",
			input:    []interface{}{nil},
			expected: nil,
		},
		{
			name: "defaults",
			input: []interface{}{map[string]interface{}{
				"strategy":    "Rolling",
				"preferences": []interface{}{},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences:          nil,
			},
		},
		{
			name: "instance_warmup only",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"instance_warmup": "60",
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay:       nil,
					CheckpointPercentages: nil,
					InstanceWarmup:        aws.Int64(60),
					MinHealthyPercentage:  nil,
				},
			},
		},
		{
			name: "instance_warmup zero",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"instance_warmup": "0",
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay:       nil,
					CheckpointPercentages: nil,
					InstanceWarmup:        aws.Int64(0),
					MinHealthyPercentage:  nil,
				},
			},
		},
		{
			name: "instance_warmup empty string",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"instance_warmup":        "",
						"min_healthy_percentage": 80,
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay:       nil,
					CheckpointPercentages: nil,
					InstanceWarmup:        nil,
					MinHealthyPercentage:  aws.Int64(80),
				},
			},
		},
		{
			name: "min_healthy_percentage only",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"min_healthy_percentage": 80,
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay:       nil,
					CheckpointPercentages: nil,
					InstanceWarmup:        nil,
					MinHealthyPercentage:  aws.Int64(80),
				},
			},
		},
		{
			name: "preferences",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"instance_warmup":        "60",
						"min_healthy_percentage": 80,
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay:       nil,
					CheckpointPercentages: nil,
					InstanceWarmup:        aws.Int64(60),
					MinHealthyPercentage:  aws.Int64(80),
				},
			},
		},
		{
			name: "checkpoint_delay",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"checkpoint_delay": "25",
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay:       aws.Int64(25),
					CheckpointPercentages: nil,
					InstanceWarmup:        nil,
					MinHealthyPercentage:  nil,
				},
			},
		},
		{
			name: "checkpoint_delay zero",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"checkpoint_delay": "0",
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay:       aws.Int64(0),
					CheckpointPercentages: nil,
					InstanceWarmup:        nil,
					MinHealthyPercentage:  nil,
				},
			},
		},
		{
			name: "checkpoint_delay empty string",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"checkpoint_delay":       "",
						"checkpoint_percentages": []interface{}{20, 100},
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay: nil,
					CheckpointPercentages: []*int64{
						aws.Int64(20),
						aws.Int64(100),
					},
					InstanceWarmup:       nil,
					MinHealthyPercentage: nil,
				},
			},
		},
		{
			name: "checkpoint_percentages empty",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"checkpoint_percentages": []interface{}{},
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay:       nil,
					CheckpointPercentages: nil,
					InstanceWarmup:        nil,
					MinHealthyPercentage:  nil,
				},
			},
		},
		{
			name: "checkpoint_percentages",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"checkpoint_percentages": []interface{}{1, 20, 25, 50, 100},
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					CheckpointDelay: nil,
					CheckpointPercentages: []*int64{
						aws.Int64(1),
						aws.Int64(20),
						aws.Int64(25),
						aws.Int64(50),
						aws.Int64(100),
					},
					InstanceWarmup:       nil,
					MinHealthyPercentage: nil,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := tfautoscaling.CreateGroupInstanceRefreshInput(asgName, testCase.input)

			if !reflect.DeepEqual(got, testCase.expected) {
				t.Errorf("got %s, expected %s", got, testCase.expected)
			}
		})
	}
}

func TestPutWarmPoolInput(t *testing.T) {
	const asgName = "test-asg"
	testCases := []struct {
		name     string
		input    []interface{}
		expected *autoscaling.PutWarmPoolInput
	}{
		{
			name:     "empty interface",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name: "only pool state",
			input: []interface{}{map[string]interface{}{
				"pool_state": "Stopped",
			}},
			expected: &autoscaling.PutWarmPoolInput{
				AutoScalingGroupName: aws.String(asgName),
				PoolState:            aws.String("Stopped"),
			},
		},
		{
			name: "0 min size",
			input: []interface{}{map[string]interface{}{
				"pool_state":                  "Stopped",
				"min_size":                    0,
				"max_group_prepared_capacity": 5,
			}},
			expected: &autoscaling.PutWarmPoolInput{
				AutoScalingGroupName:     aws.String(asgName),
				PoolState:                aws.String("Stopped"),
				MinSize:                  aws.Int64(0),
				MaxGroupPreparedCapacity: aws.Int64(5),
			},
		},
		{
			name: "-1 max prepared size",
			input: []interface{}{map[string]interface{}{
				"pool_state":                  "Stopped",
				"max_group_prepared_capacity": -1,
			}},
			expected: &autoscaling.PutWarmPoolInput{
				AutoScalingGroupName:     aws.String(asgName),
				PoolState:                aws.String("Stopped"),
				MaxGroupPreparedCapacity: aws.Int64(-1),
			},
		},
		{
			name: "all values",
			input: []interface{}{map[string]interface{}{
				"pool_state":                  "Stopped",
				"min_size":                    3,
				"max_group_prepared_capacity": 2,
			}},
			expected: &autoscaling.PutWarmPoolInput{
				AutoScalingGroupName:     aws.String(asgName),
				PoolState:                aws.String("Stopped"),
				MinSize:                  aws.Int64(3),
				MaxGroupPreparedCapacity: aws.Int64(2),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := tfautoscaling.CreatePutWarmPoolInput(asgName, testCase.input)

			if !reflect.DeepEqual(got, testCase.expected) {
				t.Errorf("got %s, expected %s", got, testCase.expected)
			}
		})
	}
}

func TestFlattenWarmPoolConfiguration(t *testing.T) {
	testCases := []struct {
		name     string
		input    *autoscaling.WarmPoolConfiguration
		expected []interface{}
	}{
		{
			name:     "empty interface",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name: "only pool state",
			input: &autoscaling.WarmPoolConfiguration{
				PoolState: aws.String("Stopped"),
			},
			expected: []interface{}{map[string]interface{}{
				"pool_state":                  "Stopped",
				"min_size":                    int64(0),
				"max_group_prepared_capacity": int64(-1),
			}},
		},
		{
			name: "only max group prepared capacity",
			input: &autoscaling.WarmPoolConfiguration{
				PoolState:                aws.String("Stopped"),
				MaxGroupPreparedCapacity: aws.Int64(2),
			},
			expected: []interface{}{map[string]interface{}{
				"pool_state":                  "Stopped",
				"min_size":                    int64(0),
				"max_group_prepared_capacity": int64(2),
			}},
		},
		{
			name: "all values",
			input: &autoscaling.WarmPoolConfiguration{
				PoolState:                aws.String("Stopped"),
				MinSize:                  aws.Int64(3),
				MaxGroupPreparedCapacity: aws.Int64(5),
			},
			expected: []interface{}{map[string]interface{}{
				"pool_state":                  "Stopped",
				"min_size":                    int64(3),
				"max_group_prepared_capacity": int64(5),
			}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := tfautoscaling.FlattenWarmPoolConfiguration(testCase.input)

			if !reflect.DeepEqual(got, testCase.expected) {
				t.Errorf("got %s, expected %s", got, testCase.expected)
			}
		})
	}
}

func testAccCheckLBTargetGroupExists(n string, res *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Target Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Conn

		targetGroup, err := tfelbv2.FindTargetGroupByARN(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading ELBv2 Target Group (%s): %w", rs.Primary.ID, err)
		}

		if targetGroup == nil {
			return fmt.Errorf("Target Group (%s) not found", rs.Primary.ID)
		}

		*res = *targetGroup
		return nil
	}
}
