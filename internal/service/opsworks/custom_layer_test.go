package opsworks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopsworks "github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
)

// These tests assume the existence of predefined Opsworks IAM roles named `aws-opsworks-ec2-role`
// and `aws-opsworks-service-role`, and Opsworks stacks named `test`.

func TestAccOpsWorksCustomLayer_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_vpcCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", "true"),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "golang"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						"type":            "gp2",
						"number_of_disks": "2",
						"mount_point":     "/home",
						"size":            "100",
						"encrypted":       "false",
					}),
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

func TestAccOpsWorksCustomLayer_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomLayerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCustomLayerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccOpsWorksCustomLayer_noVPC(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_noVPCCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", "true"),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "golang"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						"type":            "gp2",
						"number_of_disks": "2",
						"mount_point":     "/home",
						"size":            "100",
						"encrypted":       "false",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomLayerConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "120"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "golang"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "subversion"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						"type":            "gp2",
						"number_of_disks": "2",
						"mount_point":     "/home",
						"size":            "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						"type":            "io1",
						"number_of_disks": "4",
						"mount_point":     "/var",
						"size":            "100",
						"raid_level":      "1",
						"iops":            "3000",
						"encrypted":       "true",
					}),
					resource.TestCheckResourceAttr(resourceName, "custom_json", `{"layer_key":"layer_value2"}`),
				),
			},
		},
	})
}

func TestAccOpsWorksCustomLayer_cloudwatch(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_cloudWatch(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_configuration.0.log_streams.0.log_group_name", logGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file", "/var/log/system.log*"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomLayerConfig_cloudWatch(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_configuration.0.log_streams.0.log_group_name", logGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file", "/var/log/system.log*"),
				),
			},
			{
				Config: testAccCustomLayerConfig_cloudWatchFull(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_configuration.0.log_streams.0.log_group_name", logGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file", "/var/log/system.lo*"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.batch_count", "2000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.batch_size", "50000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.buffer_duration", "6000"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.encoding", "mac_turkish"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.file_fingerprint_lines", "2"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.initial_position", "end_of_file"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.multiline_start_pattern", "test*"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_configuration.0.log_streams.0.time_zone", "LOCAL"),
				),
			},
		},
	})
}

func TestAccOpsWorksCustomLayer_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerConfig_noVPCCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					acctest.CheckResourceDisappears(acctest.Provider, tfopsworks.ResourceCustomLayer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLayerExists(n string, opslayer *opsworks.Layer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		params := &opsworks.DescribeLayersInput{
			LayerIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeLayers(params)

		if err != nil {
			return err
		}

		if v := len(resp.Layers); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		*opslayer = *resp.Layers[0]

		return nil
	}
}

func testAccCheckLayerDestroy(resourceType string, s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}
		req := &opsworks.DescribeLayersInput{
			LayerIds: []*string{
				aws.String(rs.Primary.ID),
			},
		}

		output, err := conn.DescribeLayers(req)
		if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		if output == nil || len(output.Layers) == 0 {
			return nil
		}

		return fmt.Errorf("OpsWorks layer %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCustomLayerDestroy(s *terraform.State) error {
	return testAccCheckLayerDestroy("aws_opsworks_custom_layer", s)
}

func testAccCustomLayerSecurityGroups(name string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-layer1" {
  name = "%[1]s-layer1"

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-layer2" {
  name = "%[1]s-layer2"

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, name)
}

func testAccCustomLayerConfig_noVPCCreate(name string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_noVPCCreate(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
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
    encrypted       = false
  }
}
`, name))
}

func testAccCustomLayerConfig_vpcCreate(name string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_noVPCCreate(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = false

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

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
`, name))
}

func testAccCustomLayerConfig_update(name string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_noVPCCreate(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-layer3" {
  name = "tf-ops-acc-layer-%[1]s"

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
    aws_security_group.tf-ops-acc-layer3.id,
  ]
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

  custom_json = <<EOF
{
  "layer_key": "layer_value2"
}
EOF
}
`, name))
}

func testAccCustomLayerConfig_tags1(name, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = false

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

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

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1))
}

func testAccCustomLayerConfig_tags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = false

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCustomLayerConfig_cloudWatch(name string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccStackConfig_noVPCCreate(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
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
`, name, enabled))
}

func testAccCustomLayerConfig_cloudWatchFull(name string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_noVPCCreate(name),
		testAccCustomLayerSecurityGroups(name),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.test.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
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
`, name))
}
