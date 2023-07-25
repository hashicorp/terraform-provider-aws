// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfopsworks "github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
)

func TestAccOpsWorksRailsAppLayer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRailsAppLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
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
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRailsAppLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopsworks.ResourceRailsAppLayer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpsWorksRailsAppLayer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRailsAppLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccRailsAppLayerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRailsAppLayerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccOpsWorksRailsAppLayer_tagsAlternateRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID)
			// This test requires a very particular AWS Region configuration
			// in order to exercise the OpsWorks classic endpoint functionality.
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckAlternateRegionIs(t, endpoints.UsWest1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckRailsAppLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_tags1AlternateRegion(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccRailsAppLayerConfig_tags2AlternateRegion(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRailsAppLayerConfig_tags1AlternateRegion(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccOpsWorksRailsAppLayer_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRailsAppLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_allAttributes(rName, "nginx_unicorn", "1.12.5", false, "4.0.60", "2.6", "2.5.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "app_server", "nginx_unicorn"),
					resource.TestCheckResourceAttr(resourceName, "bundler_version", "1.12.5"),
					resource.TestCheckResourceAttr(resourceName, "manage_bundler", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "passenger_version", "4.0.60"),
					resource.TestCheckResourceAttr(resourceName, "ruby_version", "2.6"),
					resource.TestCheckResourceAttr(resourceName, "rubygems_version", "2.5.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRailsAppLayerConfig_allAttributes(rName, "apache_passenger", "1.15.4", true, "5.1.3", "2.3", "2.7.9"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "app_server", "apache_passenger"),
					resource.TestCheckResourceAttr(resourceName, "bundler_version", "1.15.4"),
					resource.TestCheckResourceAttr(resourceName, "manage_bundler", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "passenger_version", "5.1.3"),
					resource.TestCheckResourceAttr(resourceName, "ruby_version", "2.3"),
					resource.TestCheckResourceAttr(resourceName, "rubygems_version", "2.7.9"),
				),
			},
		},
	})
}

func TestAccOpsWorksRailsAppLayer_elb(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_rails_app_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRailsAppLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRailsAppLayerConfig_elb(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "elastic_load_balancer", "aws_elb.test.0", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRailsAppLayerConfig_elb(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "elastic_load_balancer", "aws_elb.test.1", "name"),
				),
			},
		},
	})
}

func testAccCheckRailsAppLayerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckLayerDestroy(ctx, "aws_opsworks_rails_app_layer", s)
	}
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

func testAccRailsAppLayerConfig_tags1AlternateRegion(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccLayerConfig_baseAlternateRegion(rName), fmt.Sprintf(`
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

func testAccRailsAppLayerConfig_tags2AlternateRegion(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccLayerConfig_baseAlternateRegion(rName), fmt.Sprintf(`
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

func testAccRailsAppLayerConfig_allAttributes(rName, appServer, bundlerVersion string, manageBundler bool, passengerVersion, rubyVersion, rubyGemsVersion string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_opsworks_rails_app_layer" "test" {
  name     = %[1]q
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = aws_security_group.test[*].id

  app_server        = %[2]q
  bundler_version   = %[3]q
  manage_bundler    = %[4]t
  passenger_version = %[5]q
  ruby_version      = %[6]q
  rubygems_version  = %[7]q
}
`, rName, appServer, bundlerVersion, manageBundler, passengerVersion, rubyVersion, rubyGemsVersion))
}

func testAccRailsAppLayerConfig_elb(rName string, idx int) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_elb" "test" {
  count = 2

  subnets = aws_subnet.test[*].id

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_opsworks_rails_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id
  name     = %[1]q

  custom_security_group_ids = aws_security_group.test[*].id

  elastic_load_balancer = aws_elb.test[%[2]d].name
}
`, rName, idx))
}
