// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopsworks "github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpsWorksInstance_basic(t *testing.T) {
	acctest.Skip(t, "skipping test; Amazon OpsWorks has been deprecated and will be removed in the next major release")

	ctx := acctest.Context(t)
	var opsinst awstypes.Instance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_instance.test"
	dataSourceName := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.OpsWorks) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &opsinst),
					testAccCheckInstanceAttributes(&opsinst),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "stopped"),
					resource.TestCheckResourceAttr(resourceName, "layer_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "install_updates_on_boot", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "default"),
					resource.TestCheckResourceAttr(resourceName, "os", "Amazon Linux 2016.09"),                              // inherited from opsworks_stack_test
					resource.TestCheckResourceAttr(resourceName, "root_device_type", "ebs"),                                 // inherited from opsworks_stack_test
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, dataSourceName, "names.0"), // inherited from opsworks_stack_test
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrState}, //state is something we pass to the API and get back as status :(
			},
			{
				Config: testAccInstanceConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &opsinst),
					testAccCheckInstanceAttributes(&opsinst),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.small"),
					resource.TestCheckResourceAttr(resourceName, "layer_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "os", "Amazon Linux 2015.09"),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "default"),
				),
			},
		},
	})
}

func TestAccOpsWorksInstance_updateHostNameForceNew(t *testing.T) {
	acctest.Skip(t, "skipping test; Amazon OpsWorks has been deprecated and will be removed in the next major release")

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_instance.test"
	var before, after awstypes.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.OpsWorks) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpsWorksServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrState},
			},
			{
				Config: testAccInstanceConfig_updateHostName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc2"),
					testAccCheckInstanceRecreated(t, &before, &after),
				),
			},
		},
	})
}

func testAccCheckInstanceRecreated(t *testing.T, before, after *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.InstanceId == *after.InstanceId {
			t.Fatalf("Expected change of OpsWorks Instance IDs, but both were %s", *before.InstanceId)
		}
		return nil
	}
}

func testAccCheckInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opsworks_instance" {
				continue
			}
			_, err := tfopsworks.FindInstanceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpsWorks Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInstanceExists(ctx context.Context, n string, opsinst *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Opsworks Instance is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksClient(ctx)

		output, err := tfopsworks.FindInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*opsinst = *output

		return nil
	}
}

func testAccCheckInstanceAttributes(opsinst *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Depending on the timing, the state could be requested or stopped
		if aws.ToString(opsinst.Status) != "stopped" && aws.ToString(opsinst.Status) != "requested" {
			return fmt.Errorf("Unexpected request status: %s", *opsinst.Status)
		}
		if opsinst.Architecture != awstypes.ArchitectureX8664 {
			return fmt.Errorf("Unexpected architecture: %s", opsinst.Architecture)
		}
		if *opsinst.Tenancy != "default" {
			return fmt.Errorf("Unexpected tenancy: %s", *opsinst.Tenancy)
		}
		if aws.ToString(opsinst.InfrastructureClass) != "ec2" {
			return fmt.Errorf("Unexpected infrastructure class: %s", aws.ToString(opsinst.InfrastructureClass))
		}
		if opsinst.RootDeviceType != awstypes.RootDeviceTypeEbs {
			return fmt.Errorf("Unexpected root device type: %s", opsinst.RootDeviceType)
		}
		if opsinst.VirtualizationType != awstypes.VirtualizationTypeHvm {
			return fmt.Errorf("Unexpected virtualization type: %s", opsinst.VirtualizationType)
		}
		return nil
	}
}

func testAccInstanceConfig_updateHostName(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%[1]s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%[1]s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  layer_ids = [
    aws_opsworks_static_web_layer.test.id,
  ]

  instance_type = "t2.micro"
  state         = "stopped"
  hostname      = "tf-acc2"
}
`, rName))
}

func testAccInstanceConfig_create(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%[1]s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%[1]s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  layer_ids = [
    aws_opsworks_static_web_layer.test.id,
  ]

  instance_type = "t2.micro"
  state         = "stopped"
  hostname      = "tf-acc1"
}
`, rName))
}

func testAccInstanceConfig_update(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%[1]s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%[1]s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  layer_ids = [
    aws_opsworks_static_web_layer.test.id,
    aws_opsworks_php_app_layer.test.id,
  ]

  instance_type = "t2.small"
  state         = "stopped"
  hostname      = "tf-acc1"
  os            = "Amazon Linux 2015.09"

  timeouts {
    update = "15s"
  }
}
`, rName))
}
