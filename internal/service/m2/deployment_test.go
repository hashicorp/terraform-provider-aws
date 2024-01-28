// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfm2 "github.com/hashicorp/terraform-provider-aws/internal/service/m2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"
	var deployment awstypes.DeploymentSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckDeploymentDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_type"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
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

func TestAccDeployment_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"
	var deployment awstypes.DeploymentSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckDeploymentDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_type"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func testAccCheckDeploymentExists(ctx context.Context, resourceName string, v *awstypes.DeploymentSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no M2 Deployment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)
		out, err := tfm2.FindDeploymentByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("retrieving M2 Deployment (%s): %w", rs.Primary.ID, err)
		}

		v = out

		return nil
	}
}

func testAccCheckDeploymentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_m2_deployment" {
				continue
			}

			_, err := tfm2.FindDeploymentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("M2 Deployment (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDeploymentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_m2_deployment" "test" {
  engine_type   = "microfocus"
  instance_type = "M2.m5.large"
  name          = %[1]q
}
`, rName)
}

func testAccDeploymentConfig_full(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  description     = "Test-1"
  engine_type     = "microfocus"
  engine_version = "8.0.10"
  high_availability_config {
	desired_capacity = 1
  }
  instance_type   = "M2.m5.large"
  kms_key_id      = aws_kms_key.test.arn
  name            = %[1]q
  security_group_ids   = [aws_security_group.test.id]
  subnet_ids           = aws_subnet.test[*].id
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  tags = {
	  Name = %[1]q
  }
}

resource "aws_kms_key" "test" {
  description = "tf-test-cmk-kms-key-id"
}
  
resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id
  
  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
	cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName))
}
