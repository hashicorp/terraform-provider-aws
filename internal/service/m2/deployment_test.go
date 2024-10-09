// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/m2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfm2 "github.com/hashicorp/terraform-provider-aws/internal/service/m2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccM2Deployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var deployment m2.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", acctest.Ct1),
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

func TestAccM2Deployment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var deployment m2.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfm2.ResourceDeployment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccM2Deployment_start(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var deployment m2.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccM2Deployment_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var deployment m2.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", acctest.Ct1),
				),
			},
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 2, 2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", acctest.Ct2),
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

func testAccCheckDeploymentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_m2_deployment" {
				continue
			}

			_, err := tfm2.FindDeploymentByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["deployment_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Mainframe Modernization Deployment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDeploymentExists(ctx context.Context, n string, v *m2.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		output, err := tfm2.FindDeploymentByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["deployment_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDeploymentConfig_basic(rName, engineType string, appVersion, deployVersion int, start bool) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), testAccApplicationConfig_versioned(rName, engineType, appVersion, 2), fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  name          = %[1]q
  engine_type   = %[2]q
  instance_type = "M2.m5.large"

  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test[*].id
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "secretsmanager" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.secretsmanager"
  vpc_endpoint_type = "Interface"

  security_group_ids = [
    aws_security_group.test.id,
  ]
  subnet_ids = aws_subnet.test[*].id

  private_dns_enabled = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_m2_deployment" "test" {
  environment_id      = aws_m2_environment.test.id
  application_id      = aws_m2_application.test.id
  application_version = %[3]d
  start               = %[4]t
  depends_on          = [aws_vpc_endpoint.secretsmanager]
}
`, rName, engineType, deployVersion, start))
}
