package opsworks_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfopsworks "github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
)

func TestAccOpsWorksRailsAppLayer_basic(t *testing.T) {
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRailsAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "app_server", "apache_passenger"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "opsworks", regexp.MustCompile(`layer/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_public_ips", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", "true"),
					resource.TestCheckResourceAttr(resourceName, "bundler_version", "1.5.3"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_configure_recipes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_deploy_recipes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_instance_profile_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_json", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "custom_setup_recipes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_shutdown_recipes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_undeploy_recipes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elastic_load_balancer", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "120"),
					resource.TestCheckResourceAttr(resourceName, "install_updates_on_boot", "true"),
					resource.TestCheckResourceAttr(resourceName, "manage_bundler", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", "Rails App Server"),
					resource.TestCheckResourceAttr(resourceName, "passenger_version", "4.0.46"),
					resource.TestCheckResourceAttr(resourceName, "ruby_version", "2.0.0"),
					resource.TestCheckResourceAttr(resourceName, "rubygems_version", "2.2.2"),
					resource.TestCheckNoResourceAttr(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "use_ebs_optimized_instances", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpsWorksRailsAppLayer_disappears(t *testing.T) {
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRailsAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfopsworks.ResourceRailsAppLayer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpsWorksRailsAppLayer_tags(t *testing.T) {
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRailsAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccRailsAppLayerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRailsAppLayerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckRailsAppLayerDestroy(s *terraform.State) error {
	return testAccCheckLayerDestroy("aws_opsworks_rails_app_layer", s)
}

func testAccRailsAppLayerConfig_vpcCreate(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		testAccCustomLayerSecurityGroups(rName),
		fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id
  name     = %[1]q

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
}
`, rName))
}

func testAccRailsAppLayerConfig_noManageBundlerVPCCreate(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		testAccCustomLayerSecurityGroups(rName),
		fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id
  name     = %[1]q

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  manage_bundler = false
}
`, rName))
}

func testAccRailsAppLayerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), `
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = aws_security_group.test[*].id
}
`)
}

func testAccRailsAppLayerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id
  name     = %[1]q

  custom_security_group_ids = aws_security_group.test[*].id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccRailsAppLayerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id
  name     = %[1]q

  custom_security_group_ids = aws_security_group.test[*].id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
