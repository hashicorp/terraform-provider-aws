// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpsWorksCustomLayer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_custom_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "opsworks", regexache.MustCompile(`layer/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_public_ips", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_configure_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_deploy_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_instance_profile_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_json", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "custom_setup_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_shutdown_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_undeploy_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						names.AttrType:      "gp2",
						"number_of_disks":   acctest.Ct2,
						"mount_point":       "/home",
						names.AttrSize:      "100",
						names.AttrEncrypted: acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "elastic_load_balancer", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "install_updates_on_boot", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "short_name", "tf-ops-acc-custom-layer"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "golang"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "use_ebs_optimized_instances", acctest.CtFalse),
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

// _disappears and _tags for OpsWorks Layers are tested via aws_opsworks_rails_app_layer.

func TestAccOpsWorksCustomLayer_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_custom_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomLayerConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "opsworks", regexache.MustCompile(`layer/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_public_ips", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_configure_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_deploy_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_instance_profile_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_json", testAccCustomJSON1),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "custom_setup_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_shutdown_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "custom_undeploy_recipes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						names.AttrType:    "gp2",
						"number_of_disks": acctest.Ct2,
						"mount_point":     "/home",
						names.AttrSize:    "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						names.AttrType:      "io1",
						"number_of_disks":   acctest.Ct4,
						"mount_point":       "/var",
						names.AttrSize:      "100",
						"raid_level":        acctest.Ct1,
						names.AttrIOPS:      "3000",
						names.AttrEncrypted: acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "elastic_load_balancer", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "120"),
					resource.TestCheckResourceAttr(resourceName, "install_updates_on_boot", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "short_name", "tf-ops-acc-custom-layer"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "golang"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "subversion"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "use_ebs_optimized_instances", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccOpsWorksCustomLayer_cloudWatch(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_custom_layer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_cloudWatch(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.batch_count", "1000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.batch_size", "32768"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.buffer_duration", "5000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.datetime_format", ""),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.encoding", "utf_8"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file", "/var/log/system.log*"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file_fingerprint_lines", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.initial_position", "start_of_file"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_configuration.0.log_streams.0.log_group_name", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.multiline_start_pattern", ""),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.time_zone", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomLayerConfig_cloudWatch(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.batch_count", "1000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.batch_size", "32768"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.buffer_duration", "5000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.datetime_format", ""),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.encoding", "utf_8"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file", "/var/log/system.log*"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file_fingerprint_lines", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.initial_position", "start_of_file"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_configuration.0.log_streams.0.log_group_name", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.multiline_start_pattern", ""),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.time_zone", ""),
				),
			},
			{
				Config: testAccCustomLayerConfig_cloudWatchFull(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.batch_count", "2000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.batch_size", "50000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.buffer_duration", "6000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.encoding", "mac_turkish"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file", "/var/log/system.lo*"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file_fingerprint_lines", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.initial_position", "end_of_file"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_configuration.0.log_streams.0.log_group_name", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.multiline_start_pattern", "test*"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.time_zone", "LOCAL"),
				),
			},
		},
	})
}

func TestAccOpsWorksCustomLayer_loadBasedAutoScaling(t *testing.T) {
	ctx := acctest.Context(t)
	var v opsworks.Layer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_custom_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, opsworks.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLayerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_loadBasedAutoScaling(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.alarms.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.cpu_threshold", "20"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.ignore_metrics_time", "15"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.instance_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.load_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.memory_threshold", "20"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.thresholds_wait_time", "30"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.alarms.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.cpu_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.ignore_metrics_time", "15"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.instance_count", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.load_threshold", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.memory_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.thresholds_wait_time", "35"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomLayerConfig_loadBasedAutoScaling(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLayerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.alarms.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.cpu_threshold", "20"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.ignore_metrics_time", "15"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.instance_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.load_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.memory_threshold", "20"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.downscaling.0.thresholds_wait_time", "30"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.alarms.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.cpu_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.ignore_metrics_time", "15"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.instance_count", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.load_threshold", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.memory_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "load_based_auto_scaling.0.upscaling.0.thresholds_wait_time", "35"),
				),
			},
		},
	})
}

func testAccCheckCustomLayerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error { return testAccCheckLayerDestroy(ctx, "aws_opsworks_custom_layer", s) }
}

func testAccCustomLayerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = false

  custom_security_group_ids = aws_security_group.test[*].id

  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300

  system_packages = [
    "git",
    "golang",
  ]

  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
  }
}
`, rName))
}

func testAccCustomLayerConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "extra" {
  name   = "%[1]s-extra"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true

  custom_security_group_ids = concat(aws_security_group.test[*].id, [aws_security_group.extra.id])

  drain_elb_on_shutdown     = false
  instance_shutdown_timeout = 120

  system_packages = [
    "git",
    "golang",
    "subversion",
  ]

  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
    encrypted       = true
  }

  ebs_volume {
    type            = "io1"
    number_of_disks = 4
    mount_point     = "/var"
    size            = 100
    raid_level      = 1
    iops            = 3000
    encrypted       = true
  }

  custom_json = %[2]q
}
`, rName, testAccCustomJSON1))
}

func testAccCustomLayerConfig_cloudWatch(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true

  custom_security_group_ids = aws_security_group.test[*].id

  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300

  cloudwatch_configuration {
    enabled = %[2]t

    log_streams {
      log_group_name = aws_cloudwatch_log_group.test.name
      file           = "/var/log/system.log*"
    }
  }
}
`, rName, enabled))
}

func testAccCustomLayerConfig_cloudWatchFull(rName string) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true

  custom_security_group_ids = aws_security_group.test[*].id

  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300

  cloudwatch_configuration {
    enabled = true

    log_streams {
      log_group_name          = aws_cloudwatch_log_group.test.name
      file                    = "/var/log/system.lo*"
      batch_count             = 2000
      batch_size              = 50000
      buffer_duration         = 6000
      encoding                = "mac_turkish"
      file_fingerprint_lines  = "2"
      initial_position        = "end_of_file"
      multiline_start_pattern = "test*"
      time_zone               = "LOCAL"
    }
  }
}
`, rName))
}

func testAccCustomLayerConfig_loadBasedAutoScaling(rName string, enable bool) string {
	return acctest.ConfigCompose(testAccLayerConfig_base(rName), fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true

  custom_security_group_ids = aws_security_group.test[*].id

  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300

  load_based_auto_scaling {
    enable = %[2]t

    downscaling {
      cpu_threshold        = 20
      ignore_metrics_time  = 15
      instance_count       = 2
      load_threshold       = 5
      memory_threshold     = 20
      thresholds_wait_time = 30
    }

    upscaling {
      cpu_threshold        = 80
      ignore_metrics_time  = 15
      instance_count       = 3
      load_threshold       = 10
      memory_threshold     = 80
      thresholds_wait_time = 35
    }
  }
}
`, rName, enable))
}
